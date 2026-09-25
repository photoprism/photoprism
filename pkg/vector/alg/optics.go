package alg

import (
	"math"
)

// Ordering is the reachability ordering an OPTICS pass produces, from which clusters are extracted
// afterwards. It is not itself a clustering: the same ordering yields different clusters depending
// on the extraction, which is the property DBSCAN lacks.
type Ordering struct {
	// Order lists the data point indices in the order OPTICS processed them.
	Order []int
	// Reachability holds each point's reachability distance, indexed by data point rather than by
	// position in Order, and Inf for the point that started a processing run.
	Reachability []float64
	// CoreDist holds each point's core distance, Inf where it has too few neighbors to be a core.
	CoreDist []float64
	// Predecessor names the point each reachability was measured from, or -1 where none was.
	Predecessor []int
	// MinPts is the core size the pass ran with, which every extraction inherits.
	MinPts int
	// MaxEps bounds the distances the pass considered, so no extraction above it is meaningful.
	MaxEps float64
}

// OPTICS orders points by reachability, so that clusters spanning a range of densities can be
// extracted from a single pass. See Ankerst et al., "OPTICS: Ordering Points To Identify the
// Clustering Structure" (1999). Pass Inf for maxEps to leave the ordering unbounded.
func OPTICS(data [][]float64, minPts int, maxEps float64, workers int, distance DistFunc) (*Ordering, error) {
	if _, err := dataDims(data); err != nil {
		return nil, err
	}

	switch {
	case minPts < 1:
		return nil, errZeroMinpts
	case workers < 0:
		return nil, errZeroWorkers
	case maxEps <= 0:
		return nil, errZeroEpsilon
	}

	if distance == nil {
		distance = EuclideanDist
	}

	n := len(data)

	o := &Ordering{
		Order:        make([]int, 0, n),
		Reachability: make([]float64, n),
		CoreDist:     coreDistances(data, minPts, maxEps, workers, distance),
		Predecessor:  make([]int, n),
		MinPts:       minPts,
		MaxEps:       maxEps,
	}

	for i := range n {
		o.Reachability[i] = math.Inf(1)
		o.Predecessor[i] = -1
	}

	processed := make([]bool, n)

	// Distances to one point, reused across expansions. A full matrix would cost n^2 floats and
	// save nothing, since each expansion recomputes one row either way.
	row := make([]float64, n)
	seeds := newReachHeap(n)

	for i := range n {
		if processed[i] {
			continue
		}

		processed[i] = true
		o.Order = append(o.Order, i)

		if math.IsInf(o.CoreDist[i], 1) {
			continue
		}

		seeds.reset()
		o.expand(data, i, processed, row, seeds, workers, distance)

		for seeds.Len() > 0 {
			q := seeds.pop()

			if processed[q] {
				continue
			}

			processed[q] = true
			o.Order = append(o.Order, q)

			if !math.IsInf(o.CoreDist[q], 1) {
				o.expand(data, q, processed, row, seeds, workers, distance)
			}
		}
	}

	return o, nil
}

// expand updates the reachability of every unprocessed neighbor of p and queues it, which is what
// makes the ordering follow density rather than input order.
func (o *Ordering) expand(data [][]float64, p int, processed []bool, row []float64, seeds *reachHeap, workers int, distance DistFunc) {
	n := len(data)

	parallelRows(n, workers, func(j int) {
		row[j] = distance(data[p], data[j])
	})

	core := o.CoreDist[p]

	for j := range n {
		if processed[j] || j == p {
			continue
		}

		d := row[j]

		if math.IsNaN(d) || d > o.MaxEps {
			continue
		}

		// A neighbor is never reachable more closely than the density around p allows, which is
		// what keeps a sparse region from reading as dense because one of its points is near.
		if r := math.Max(core, d); r < o.Reachability[j] {
			o.Reachability[j] = r
			o.Predecessor[j] = p
			seeds.push(j, r)
		}
	}
}

// Plot returns the reachability distances in processing order, which is the curve clusters appear
// in as valleys. The leading Inf is kept, so the result aligns with Order index by index.
func (o *Ordering) Plot() []float64 {
	plot := make([]float64, len(o.Order))

	for i, p := range o.Order {
		plot[i] = o.Reachability[p]
	}

	return plot
}

// ExtractDBSCAN returns the textbook DBSCAN clustering at the given link distance, which one
// ordering already contains for every eps up to MaxEps. Reported for comparison and as a control.
//
// It agrees with this package's DBSCAN on the core points. The two differ on border points, which
// this extraction takes in the order the ordering reached them and DBSCAN attaches only where the
// cores around them agree.
func (o *Ordering) ExtractDBSCAN(eps float64) Labels {
	labels := o.noiseLabels()
	current := 0

	for _, p := range o.Order {
		if o.Reachability[p] > eps {
			// A point too far to be reached starts a cluster when it is dense enough itself, and
			// is noise otherwise, which also ends the cluster before it.
			if o.CoreDist[p] <= eps {
				current++
				labels[p] = current
			}
		} else if current > 0 {
			labels[p] = current
		}
	}

	return labels
}

// ExtractXi returns the clusters that appear as valleys in the reachability plot, each bounded by
// a steep fall and a later steep rise. See Ankerst et al. § 4.2, with the corrections in
// scikit-learn's cluster_optics_xi.
//
// xi is the relative height a wall must have, so one value spans densities no single link distance
// covers. minSize is the smallest cluster returned, and defaults to MinPts when below it. Where
// valleys nest, the innermost is returned, which is what splits a merged group into its people.
func (o *Ordering) ExtractXi(xi float64, minSize int) Labels {
	if xi <= 0 || xi >= 1 || len(o.Order) == 0 {
		return o.noiseLabels()
	}

	if minSize < o.MinPts {
		minSize = o.MinPts
	}

	return o.labelRanges(o.xiRanges(xi, minSize), minSize)
}

// xiRange is a cluster found in the plot, as an inclusive span of positions in Order.
type xiRange struct {
	start, end int
}

// steepAreaSet is the state the plot walk carries: the falls still open, and the highest
// reachability seen since each was recorded.
type steepArea struct {
	start, end int
	mib        float64
}

// xiRanges walks the reachability plot once and returns every valley bounded by a steep fall and
// a steep rise, innermost first within each rise.
func (o *Ordering) xiRanges(xi float64, minSize int) []xiRange {
	// A sentinel above every real value, so that a cluster ending at the last position still has
	// a following height to be compared against.
	plot := append(o.Plot(), math.Inf(1))
	n := len(plot) - 1

	comp := 1 - xi

	steepUp := make([]bool, n)
	steepDown := make([]bool, n)
	up := make([]bool, n)
	down := make([]bool, n)

	for i := range n {
		// Inf over Inf is NaN, and every comparison against it is false, which correctly leaves
		// a run of unreachable points outside any wall.
		ratio := plot[i] / plot[i+1]

		steepUp[i] = ratio <= comp
		steepDown[i] = ratio >= 1/comp
		up[i] = ratio < 1
		down[i] = ratio > 1
	}

	var (
		areas  []steepArea
		ranges []xiRange
	)

	index, mib := 0, 0.0

	for i := range n {
		if !steepUp[i] && !steepDown[i] || i < index {
			continue
		}

		for j := index; j <= i; j++ {
			mib = math.Max(mib, plot[j])
		}

		areas = filterAreas(areas, plot, mib, comp)

		if steepDown[i] {
			end := extendArea(steepDown, up, i, o.MinPts)
			areas = append(areas, steepArea{start: i, end: end})

			index = end + 1
			mib = plot[index]

			continue
		}

		upStart := i
		upEnd := extendArea(steepUp, down, i, o.MinPts)

		index = upEnd + 1
		mib = plot[index]

		ranges = append(ranges, o.pairAreas(plot, areas, upStart, upEnd, comp, minSize)...)
	}

	return ranges
}

// filterAreas drops the falls that the plot has since climbed back over, and carries that maximum
// onto the ones that remain. A fall the reachability has risen past can no longer bound a cluster,
// because the points after it are not denser than the ones before.
func filterAreas(areas []steepArea, plot []float64, mib, comp float64) []steepArea {
	if math.IsInf(mib, 1) {
		return nil
	}

	kept := areas[:0]

	for _, a := range areas {
		if mib > plot[a.start]*comp {
			continue
		}

		a.mib = math.Max(a.mib, mib)
		kept = append(kept, a)
	}

	return kept
}

// extendArea returns the last position of the wall starting at start, absorbing up to minPts
// consecutive points that are not steep but still continue in the same direction. A wall broken by
// a short plateau is one wall, and treating it as two would split a cluster at the plateau.
func extendArea(steep, against []bool, start, minPts int) int {
	end, flat := start, 0

	for i := start; i < len(steep); i++ {
		switch {
		case steep[i]:
			end, flat = i, 0
		case !against[i]:
			if flat++; flat > minPts {
				return end
			}
		default:
			return end
		}
	}

	return end
}

// pairAreas returns the valleys formed by pairing each open fall with the rise that closes it,
// applying the height and size conditions a xi-steep cluster has to satisfy. Innermost first, so
// that labelRanges keeps the children of a merged group rather than the group.
func (o *Ordering) pairAreas(plot []float64, areas []steepArea, upStart, upEnd int, comp float64, minSize int) []xiRange {
	found := make([]xiRange, 0, len(areas))

	for _, down := range areas {
		start, end := down.start, upEnd

		// The rise has to clear the highest point since the fall, or the two bound a step between
		// densities rather than one valley.
		if plot[end+1]*comp < down.mib {
			continue
		}

		// The walls have to reach comparable heights. Where one is taller, the cluster begins or
		// ends partway along it, at the level the shorter one reaches.
		if plot[down.start]*comp >= plot[end+1] {
			for start < down.end && plot[start+1] > plot[end+1] {
				start++
			}
		} else if plot[end+1]*comp >= plot[down.start] {
			// Above rather than below: the walk trims the rise back to where it last stood under
			// the height the fall started at, which is the level the two walls share.
			for end > upStart && plot[end-1] > plot[down.start] {
				end--
			}
		}

		start, end = o.correctPredecessor(plot, start, end)

		if start < 0 || end-start+1 < minSize || start > down.end || end < upStart {
			continue
		}

		found = append(found, xiRange{start: start, end: end})
	}

	// Reversed, because the falls are recorded outermost first and the innermost valley is the
	// one a nested pair should return.
	for i, j := 0, len(found)-1; i < j; i, j = i+1, j-1 {
		found[i], found[j] = found[j], found[i]
	}

	return found
}

// correctPredecessor trims a valley's end until the point there was reached from inside it, which
// keeps a point that merely follows the cluster in the ordering from being counted as part of it.
// Reports a start of -1 when nothing in the span qualifies.
func (o *Ordering) correctPredecessor(plot []float64, start, end int) (int, int) {
	for start < end {
		if plot[start] > plot[end] {
			return start, end
		}

		pred := o.Predecessor[o.Order[end]]

		for i := start; i < end; i++ {
			if pred == o.Order[i] {
				return start, end
			}
		}

		end--
	}

	return -1, -1
}

// labelRanges turns the valleys into labels, keeping the first that covers each span. The ranges
// arrive innermost first, so a nested pair yields its children and the merged parent is dropped.
func (o *Ordering) labelRanges(ranges []xiRange, minSize int) Labels {
	labels := o.noiseLabels()
	taken := make([]bool, len(o.Order))
	next := 0

	for _, r := range ranges {
		if r.end-r.start+1 < minSize {
			continue
		}

		overlaps := false

		for i := r.start; i <= r.end; i++ {
			if taken[i] {
				overlaps = true
				break
			}
		}

		if overlaps {
			continue
		}

		next++

		for i := r.start; i <= r.end; i++ {
			taken[i] = true
			labels[o.Order[i]] = next
		}
	}

	return labels
}

// noiseLabels returns one label per point, all noise.
func (o *Ordering) noiseLabels() Labels {
	labels := make(Labels, len(o.Reachability))

	for i := range labels {
		labels[i] = Noise
	}

	return labels
}
