package alg

import (
	"math"
	"sort"
)

// Hierarchy is the condensed cluster tree an HDBSCAN pass produces. Clusters are selected from it
// by how long they persist as the density level falls, rather than at one link distance, so groups
// of different densities are found in a single pass.
type Hierarchy struct {
	// MinPts is the core size the pass ran with, the same quantity DBSCAN calls its core.
	MinPts int
	// MinSize is the smallest group the condensing kept as a cluster of its own.
	MinSize int
	// CoreDist holds the distance from every point to its MinPts-th nearest neighbor.
	CoreDist []float64

	n int
	// mst holds the spanning tree edges in ascending distance, which is the order the hierarchy
	// merges in and the order a cut at one distance walks.
	mst []mstEdge
	// merges is the single-linkage hierarchy over the mutual reachability distances, in merge
	// order, with node n+i naming the i-th merge.
	merges []slMerge
	// tree is the condensed hierarchy, holding one edge per point or cluster leaving its parent.
	tree []condensedEdge
	// stability sums, per condensed cluster, how long its points persisted within it.
	stability map[int]float64
	// selected names the clusters the extraction chose, and labelOf maps them to label numbers.
	selected []int
	labelOf  map[int]int

	labels        Labels
	probabilities []float64
	outliers      []float64
}

// slMerge is one join in the single-linkage hierarchy, naming the two nodes it joined, the
// mutual reachability distance it happened at, and how many points the result holds.
type slMerge struct {
	left, right int
	dist        float64
	size        int
}

// condensedEdge records a point or a cluster leaving its parent cluster, at the density level
// lambda, which is the reciprocal of the distance the split happened at.
type condensedEdge struct {
	parent, child int
	lambda        float64
	size          int
}

// HDBSCAN clusters by building the hierarchy of all density levels at once and keeping the
// clusters that persist longest across it. See Campello, Moulavi and Sander, "Density-Based
// Clustering Based on Hierarchical Density Estimates" (2013).
//
// minPts sets the core size, as in DBSCAN. minSize is the smallest group treated as a cluster of
// its own rather than as points falling out of its parent, and defaults to minPts when below it.
func HDBSCAN(data [][]float64, minPts, minSize, workers int, distance DistFunc) (*Hierarchy, error) {
	if _, err := dataDims(data); err != nil {
		return nil, err
	}

	switch {
	case minPts < 1:
		return nil, errZeroMinpts
	case workers < 0:
		return nil, errZeroWorkers
	}

	if distance == nil {
		distance = EuclideanDist
	}

	if minSize < minPts {
		minSize = minPts
	}

	n := len(data)

	h := &Hierarchy{
		MinPts:   minPts,
		MinSize:  minSize,
		CoreDist: coreDistances(data, minPts, math.Inf(1), workers, distance),
		n:        n,
		labelOf:  make(map[int]int),
	}

	// A single point has no hierarchy to condense, and no cluster can reach minSize.
	if n < 2 {
		h.labels = make(Labels, n)
		h.probabilities = make([]float64, n)
		h.outliers = make([]float64, n)

		for i := range h.labels {
			h.labels[i] = Noise
		}

		return h, nil
	}

	h.mst = h.minimumSpanningTree(data, workers, distance)
	h.merges = h.singleLinkage(h.mst)
	h.tree = h.condense()
	h.stability = h.computeStability()
	h.selected = h.selectClusters()
	h.label()

	return h, nil
}

// mstEdge is one edge of the minimum spanning tree over the mutual reachability distances.
type mstEdge struct {
	a, b int
	dist float64
}

// mutualReach returns the mutual reachability distance between two points, which is the plain
// distance widened to the sparser of the two neighborhoods.
//
// Widening is what makes a point in a sparse region expensive to reach from a dense one, so a thin
// bridge of outliers cannot chain two clusters together the way it does under a plain distance.
func (h *Hierarchy) mutualReach(a, b int, d float64) float64 {
	return math.Max(d, math.Max(h.CoreDist[a], h.CoreDist[b]))
}

// minimumSpanningTree returns the tree connecting every point at the shortest possible mutual
// reachability, using Prim's algorithm because the graph is complete and holds no explicit edges.
func (h *Hierarchy) minimumSpanningTree(data [][]float64, workers int, distance DistFunc) []mstEdge {
	n := h.n

	inTree := make([]bool, n)
	best := make([]float64, n)
	from := make([]int, n)

	for i := range n {
		best[i] = math.Inf(1)
		from[i] = -1
	}

	edges := make([]mstEdge, 0, n-1)
	current := 0
	inTree[0] = true

	for range n - 1 {
		parallelRows(n, workers, func(j int) {
			if inTree[j] {
				return
			}

			if d := h.mutualReach(current, j, distance(data[current], data[j])); d < best[j] {
				best[j] = d
				from[j] = current
			}
		})

		next, shortest := -1, math.Inf(1)

		for j := range n {
			if !inTree[j] && best[j] < shortest {
				next, shortest = j, best[j]
			}
		}

		// A distance function that reports NaN leaves points unreachable, so the tree is closed
		// with whatever remains rather than looping without progress.
		if next < 0 {
			break
		}

		inTree[next] = true
		edges = append(edges, mstEdge{a: from[next], b: next, dist: shortest})
		current = next
	}

	return edges
}

// singleLinkage turns the spanning tree into the merge sequence of the single-linkage hierarchy,
// by joining components in order of the distance that connects them.
func (h *Hierarchy) singleLinkage(edges []mstEdge) []slMerge {
	sort.SliceStable(edges, func(i, j int) bool { return edges[i].dist < edges[j].dist })

	n := h.n
	uf := newUnionFind(n)

	// node names the hierarchy node each component is currently known as, which starts as the
	// point itself and becomes the merge that last joined it.
	node := make([]int, n)
	size := make([]int, n)

	for i := range n {
		node[i] = i
		size[i] = 1
	}

	merges := make([]slMerge, 0, len(edges))

	for _, e := range edges {
		ra, rb := uf.find(e.a), uf.find(e.b)

		if ra == rb {
			continue
		}

		merged := slMerge{
			left:  node[ra],
			right: node[rb],
			dist:  e.dist,
			size:  size[ra] + size[rb],
		}

		merges = append(merges, merged)

		root := uf.union(ra, rb)
		node[root] = n + len(merges) - 1
		size[root] = merged.size
	}

	return merges
}

// nodeSize returns the number of points a hierarchy node holds.
func (h *Hierarchy) nodeSize(node int) int {
	if node < h.n {
		return 1
	}

	return h.merges[node-h.n].size
}

// descendants returns every point below a hierarchy node.
func (h *Hierarchy) descendants(node int) []int {
	points := make([]int, 0, h.nodeSize(node))
	queue := []int{node}

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		if v < h.n {
			points = append(points, v)
			continue
		}

		m := h.merges[v-h.n]
		queue = append(queue, m.left, m.right)
	}

	return points
}

// lambdaOf returns the density level a distance corresponds to. Coincident points split at an
// infinite level, since no reduction in density ever separates them.
func lambdaOf(dist float64) float64 {
	if dist <= 0 {
		return math.Inf(1)
	}

	return 1 / dist
}

// condense walks the hierarchy from the top and keeps only the splits where both sides are large
// enough to be clusters. Anything smaller is recorded as its points falling out of the parent,
// which is what turns a tree of n-1 arbitrary merges into one of real groups.
func (h *Hierarchy) condense() []condensedEdge {
	if len(h.merges) == 0 {
		return nil
	}

	root := h.n + len(h.merges) - 1

	// relabel names each surviving hierarchy node by the condensed cluster it belongs to, so that
	// a chain of splits that shed only small groups stays one cluster.
	relabel := make(map[int]int, len(h.merges))
	relabel[root] = h.n

	next := h.n + 1
	tree := make([]condensedEdge, 0, h.n)

	// Top down, so a parent is always relabeled before the children that read it. Node numbers
	// increase with merge order, so descending order is exactly that.
	for node := root; node >= h.n; node-- {
		parent, ok := relabel[node]

		if !ok {
			continue
		}

		m := h.merges[node-h.n]
		lambda := lambdaOf(m.dist)

		leftSize, rightSize := h.nodeSize(m.left), h.nodeSize(m.right)
		leftBig, rightBig := leftSize >= h.MinSize, rightSize >= h.MinSize

		switch {
		case leftBig && rightBig:
			// A genuine split: both sides become clusters in their own right.
			relabel[m.left] = next
			tree = append(tree, condensedEdge{parent: parent, child: next, lambda: lambda, size: leftSize})
			next++

			relabel[m.right] = next
			tree = append(tree, condensedEdge{parent: parent, child: next, lambda: lambda, size: rightSize})
			next++
		case leftBig:
			relabel[m.left] = parent
			tree = append(tree, h.shedEdges(parent, m.right, lambda)...)
		case rightBig:
			relabel[m.right] = parent
			tree = append(tree, h.shedEdges(parent, m.left, lambda)...)
		default:
			// Neither side can carry the cluster on, so it ends here and all of its points fall
			// out at this level.
			tree = append(tree, h.shedEdges(parent, m.left, lambda)...)
			tree = append(tree, h.shedEdges(parent, m.right, lambda)...)
		}
	}

	return tree
}

// shedEdges records every point below a node as leaving the given cluster at the given level.
func (h *Hierarchy) shedEdges(parent, node int, lambda float64) []condensedEdge {
	points := h.descendants(node)
	edges := make([]condensedEdge, 0, len(points))

	for _, p := range points {
		edges = append(edges, condensedEdge{parent: parent, child: p, lambda: lambda, size: 1})
	}

	return edges
}

// computeStability returns, per condensed cluster, the total density range over which its points
// remained in it. That is the quantity the extraction maximizes, and it is what lets a merged pair
// lose to its two halves without any distance being named.
func (h *Hierarchy) computeStability() map[int]float64 {
	births := make(map[int]float64, len(h.tree))

	for _, e := range h.tree {
		if e.child >= h.n {
			births[e.child] = e.lambda
		}
	}

	// The root exists at every level, so it is born where density is lowest.
	births[h.n] = 0

	stability := make(map[int]float64, len(births))

	for _, e := range h.tree {
		stability[e.parent] += (e.lambda - births[e.parent]) * float64(e.size)
	}

	return stability
}

// childClusters returns the condensed clusters directly below the given one.
func (h *Hierarchy) childClusters(cluster int) []int {
	var children []int

	for _, e := range h.tree {
		if e.parent == cluster && e.child >= h.n {
			children = append(children, e.child)
		}
	}

	return children
}

// selectClusters chooses the clusters to report, keeping a parent only when it persists longer
// than its children together. Excess of mass, as in Campello et al. § 4.
func (h *Hierarchy) selectClusters() []int {
	nodes := make([]int, 0, len(h.stability))

	for c := range h.stability {
		nodes = append(nodes, c)
	}

	// Descending, which is bottom up: condensing numbers a child above its parent.
	sort.Sort(sort.Reverse(sort.IntSlice(nodes)))

	isCluster := make(map[int]bool, len(nodes))

	for _, c := range nodes {
		isCluster[c] = true
	}

	root := h.n

	for _, c := range nodes {
		if c == root {
			continue
		}

		children := h.childClusters(c)

		if len(children) == 0 {
			continue
		}

		below := 0.0

		for _, child := range children {
			below += h.stability[child]
		}

		if below > h.stability[c] {
			// The children together outlast the parent, so the parent is not reported and its
			// stability becomes theirs for the comparison one level up.
			isCluster[c] = false
			h.stability[c] = below
		} else {
			for _, sub := range h.subClusters(c) {
				isCluster[sub] = false
			}
		}
	}

	// The root is only ever a cluster when nothing below it was chosen, which would report the
	// whole set as one group and say nothing.
	isCluster[root] = false

	selected := make([]int, 0, len(nodes))

	for _, c := range nodes {
		if isCluster[c] {
			selected = append(selected, c)
		}
	}

	sort.Ints(selected)

	return selected
}

// subClusters returns every condensed cluster below the given one, excluding itself.
func (h *Hierarchy) subClusters(cluster int) []int {
	var (
		found []int
		queue = []int{cluster}
	)

	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]

		for _, child := range h.childClusters(c) {
			found = append(found, child)
			queue = append(queue, child)
		}
	}

	return found
}

// label assigns every point to the deepest selected cluster holding it, and derives the per-point
// membership and outlier scores from the levels it left at.
func (h *Hierarchy) label() {
	h.labels = make(Labels, h.n)
	h.probabilities = make([]float64, h.n)
	h.outliers = make([]float64, h.n)

	for i := range h.labels {
		h.labels[i] = Noise
	}

	for i, c := range h.selected {
		h.labelOf[c] = i + 1
	}

	// Clusters that were not selected are folded into their parent, so that following a point's
	// parent upward stops at the first selected ancestor, or at the root when there is none.
	up := make(ancestry, len(h.tree))

	for _, e := range h.tree {
		if e.child >= h.n && h.labelOf[e.child] == 0 {
			up.attach(e.child, e.parent)
		}
	}

	deaths := h.clusterDeaths()

	for _, e := range h.tree {
		if e.child >= h.n {
			continue
		}

		cluster := up.root(e.parent)
		label := h.labelOf[cluster]

		if label == 0 {
			h.outliers[e.child] = outlierScore(deaths[up.root(e.parent)], e.lambda)
			continue
		}

		h.labels[e.child] = label
		h.probabilities[e.child] = membership(deaths[cluster], e.lambda)
		h.outliers[e.child] = outlierScore(deaths[cluster], e.lambda)
	}

	// A point that never left its cluster stays in it to the end, so it is a full member.
	for _, c := range h.selected {
		for _, p := range h.clusterPoints(c) {
			if h.labels[p] == Noise {
				h.labels[p] = h.labelOf[c]
				h.probabilities[p] = 1
			}
		}
	}
}

// clusterPoints returns every point that a condensed cluster or its descendants hold.
func (h *Hierarchy) clusterPoints(cluster int) []int {
	var points []int

	within := map[int]bool{cluster: true}

	for _, c := range h.subClusters(cluster) {
		within[c] = true
	}

	for _, e := range h.tree {
		if e.child < h.n && within[e.parent] {
			points = append(points, e.child)
		}
	}

	return points
}

// clusterDeaths returns the highest density level reached within each cluster, which is what a
// point's own level is measured against. The level is carried down from a parent when it is
// higher, so a sparse cluster inside a dense one is not scored against itself alone.
func (h *Hierarchy) clusterDeaths() map[int]float64 {
	deaths := make(map[int]float64, len(h.stability))
	parents := make(map[int]int, len(h.stability))

	for _, e := range h.tree {
		if e.lambda > deaths[e.parent] && !math.IsInf(e.lambda, 1) {
			deaths[e.parent] = e.lambda
		}

		if e.child >= h.n {
			parents[e.child] = e.parent
		}
	}

	clusters := make([]int, 0, len(deaths))

	for c := range deaths {
		clusters = append(clusters, c)
	}

	// Ascending is top down, so a parent's level is final before its children read it.
	sort.Ints(clusters)

	for _, c := range clusters {
		if p, ok := parents[c]; ok && deaths[p] > deaths[c] {
			deaths[c] = deaths[p]
		}
	}

	return deaths
}

// membership returns how far into its cluster's density range a point persisted, on a scale where
// one means it stayed to the end.
func membership(death, lambda float64) float64 {
	if death <= 0 || math.IsInf(lambda, 1) {
		return 1
	}

	return math.Min(lambda, death) / death
}

// outlierScore returns how early a point left its cluster relative to the level the cluster
// reached, which is the GLOSH score of Campello et al. § 5.
func outlierScore(death, lambda float64) float64 {
	if death <= 0 || math.IsInf(lambda, 1) {
		return 0
	}

	return math.Max(0, (death-lambda)/death)
}

// Labels returns the cluster each point was assigned to, numbered from one, with Noise for the
// points no cluster kept.
func (h *Hierarchy) Labels() Labels {
	return h.labels
}

// Probabilities returns how strongly each point belongs to its cluster, between zero and one, and
// zero for noise. It falls out of the same tree the labels do, rather than from a separate model.
func (h *Hierarchy) Probabilities() []float64 {
	return h.probabilities
}

// Outliers returns the GLOSH outlier score of each point, where zero is a full member and values
// approaching one mark a point that left its cluster while the cluster was still forming.
func (h *Hierarchy) Outliers() []float64 {
	return h.outliers
}

// ExtractDBSCAN returns the clustering that cutting the hierarchy at one distance produces, which
// is DBSCAN with the border points left out. Reported as a control, so an arm that beats DBSCAN
// can be compared against it on one code path.
func (h *Hierarchy) ExtractDBSCAN(eps float64) Labels {
	labels := make(Labels, h.n)

	for i := range labels {
		labels[i] = Noise
	}

	uf := newUnionFind(h.n)

	for _, e := range h.mst {
		if e.dist > eps {
			break
		}

		uf.union(e.a, e.b)
	}

	groups := make(map[int][]int)

	for i := range h.n {
		if h.CoreDist[i] > eps {
			continue
		}

		root := uf.find(i)
		groups[root] = append(groups[root], i)
	}

	roots := make([]int, 0, len(groups))

	for r := range groups {
		roots = append(roots, r)
	}

	sort.Ints(roots)

	next := 0

	for _, r := range roots {
		if len(groups[r]) < h.MinSize {
			continue
		}

		next++

		for _, p := range groups[r] {
			labels[p] = next
		}
	}

	return labels
}
