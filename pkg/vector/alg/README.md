## Package Clusters

Implements the following clustering algorithms:

- k-means++
- DBSCAN
- OPTICS, with cluster extraction at one link distance or by the xi-steep method
- HDBSCAN, with excess-of-mass extraction, membership probabilities and GLOSH outlier scores

It was forked from the following repositories, which don't seem to be maintained anymore:

- https://github.com/okhowang/clusters 
- https://github.com/mpraski/clusters

This package also provides utilities for importing data and estimating the optimal number of clusters.

### About

This library was built out of necessity for a collection of performant cluster analysis utilities for Golang. Go, thanks to its numerous advantages (single binary distribution, relative performance, growing community) seems to become an attractive alternative to languages commonly used in statistical computations and machine learning, yet it still lacks crucial tools and libraries. I use the [*floats* package](https://github.com/gonum/gonum/tree/master/floats) from the robust Gonum library to perform optimized vector calculations in tight loops.

### Installation

If you have Go 1.7+
```bash
go get github.com/photoprism/photoprism/pkg/alg
```

### Usage

The currently supported hard clustering algorithms are represented by the *HardClusterer* interface, which defines several common operations. To show an example we create, train and use a KMeans++ clusterer:

```go
var data [][]float64
var observation []float64

// Create a new KMeans++ clusterer with 1000 iterations, 
// 8 clusters and a distance measurement function of type func([]float64, []float64) float64).
// Pass nil to use clusters.EuclideanDist
c, e := clusters.KMeans(1000, 8, clusters.EuclideanDist)
if e != nil {
	panic(e)
}

// Use the data to train the clusterer
if e = c.Learn(data); e != nil {
	panic(e)
}

fmt.Printf("Clustered data set into %d\n", c.Sizes())

fmt.Printf("Assigned observation %v to cluster %d\n", observation, c.Predict(observation))

for index, number := range c.Guesses() {
	fmt.Printf("Assigned data point %v to cluster %d\n", data[index], number)
}
```

All data points must have the same number of dimensions, since vectors of different widths cannot be compared. `Learn` and `Estimate` return an error otherwise, and `Predict` returns `-1` for an observation of a different length or before the clusterer is trained.

Algorithms currently supported are KMeans++, DBSCAN, OPTICS and HDBSCAN.

### Density-Based Clustering at More Than One Density

DBSCAN takes a single link distance, so a set holding groups of different densities has no value
that suits all of them: whatever connects the sparsest group also chains the closest pair together.
`OPTICS` and `HDBSCAN` address that, and neither implements the `HardClusterer` interface, because
neither produces one clustering - they build a structure that clusterings are extracted from.

```go
// OPTICS orders the points once; clusters are then extracted from that ordering.
o, err := alg.OPTICS(data, minPts, math.Inf(1), workers, alg.EuclideanDist)

labels := o.ExtractXi(0.05, minPts) // valleys in the reachability plot
same := o.ExtractDBSCAN(0.8)        // what DBSCAN would return at this link distance

// HDBSCAN builds the hierarchy of all density levels and keeps the clusters that persist longest.
h, err := alg.HDBSCAN(data, minPts, minClusterSize, workers, alg.EuclideanDist)

labels := h.Labels()          // clusters numbered from 1, alg.Noise for the rest
probs := h.Probabilities()    // how central each point is within its own cluster
outliers := h.Outliers()      // GLOSH score, 0 for a full member
```

Both return `alg.Labels`, which numbers clusters from 1 and marks unclustered points `alg.Noise`,
matching what `HardClusterer.Guesses` reports.

### Where DBSCAN Departs From the Textbook

Border points - those inside `eps` of a core point without being cores themselves - are assigned
here, but only while they are still unvisited. A point the scan reached first was labeled noise, and
that label is final. Two behaviors follow, and callers that read the core size as a guarantee are
surprised by both:

- **Which border points get in depends on the order the points arrive in.** Not on scheduling: a
  rerun over the same slice repeats exactly. Reorder the same data and the assignment moves.
- **The core size bounds what may seed a cluster, not how large one ends up.** Expansion claims only
  neighbors that are still unassigned, so a core point reached after an earlier cluster absorbed its
  neighborhood produces a cluster smaller than `minPts` - down to the seed alone.

`Sizes()` therefore reports what each cluster claimed rather than the density that formed it. Filter
on the result if a caller needs a floor, and do not rely on the clustering being a function of the
point set alone.

Two properties worth knowing before reading a result:

- **`Probabilities` is scaled per cluster.** Every cluster reaches 1 however sparse it is, so the
  value says where a point sits within its own cluster and is not comparable between clusters.
- **Tied distances make the outcome arbitrary.** A point whose core distance exceeds every distance
  around it has several equally short edges, and which one is taken decides its cluster. Both
  implementations break such ties by point index so a run repeats, but a different valid answer
  exists.

Algorithms which support online learning can be trained this way using Online() function, which relies on channel communication to coordinate the process:

```go
c, e := clusters.KmeansClusterer(1000, 8, clusters.EuclideanDist)
if e != nil {
	panic(e)
}

c = c.WithOnline(clusters.Online{
	Alpha:     0.5,
	Dimension: 4,
})

var (
	send   = make(chan []float64)
	finish = make(chan struct{})
)

events := c.Online(send, finish)

go func() {
	for {
		select {
		case e := <-events:
			fmt.Printf("Classified observation %v into cluster: %d\n", e.Observation, e.Cluster)
		}
	}
}()

for i := 0; i < 10000; i++ {
	point := make([]float64, 4)
	for j := 0; j < 4; j++ {
		point[j] = 10 * (rand.Float64() - 0.5)
	}
	send <- point
}

finish <- struct{}{}

fmt.Printf("Clustered data set into %d\n", c.Sizes())
```

The Estimator interface defines an operation of guessing an optimal number of clusters in a dataset. As of now the KMeansEstimator is implemented using gap statistic and k-means++ as the clustering algorithm (see https://dl.photoprism.app/pdf/publications/20020106-Estimating_the_Number_of_Clusters.pdf):

```go
var data [][]float64

// Create a new KMeans++ estimator with 1000 iterations, 
// a maximum of 8 clusters and default (EuclideanDist) distance measurement
c, e := clusters.KMeansEstimator(1000, 8, clusters.EuclideanDist)
if e != nil {
	panic(e)
}

r, e := c.Estimate(data)
if e != nil {
	panic(e)
}

fmt.Printf("Estimated number of clusters: %d\n", r)

```

The library also provides an Importer to load data from file (as of now the CSV importer is implemented):

```go
// Import first three columns from data.csv
d, e := i.Import("data.csv", 0, 2)
if e != nil {
	panic(e)
}
```

### License

MIT
