package face

// Cluster represents an embedding centroid together with the maximum
// allowed distance (radius) for matches. The radius uses Euclidean distance
// between L2-normalized embeddings, so values should typically stay in the
// range [0,2].
type Cluster struct {
	Name      string    `json:"name,omitempty"`
	Radius    float64   `json:"radius,omitempty"`
	Disabled  bool      `json:"disabled,omitempty"`
	Embedding Embedding `json:"embedding,omitempty"`
}

// Clusters represents a set of clusters.
type Clusters []Cluster

// Contains reports whether the provided embedding falls inside any enabled
// cluster radius. Disabled clusters are skipped so they do not influence
// the result.
//
// Embeddings that cannot be compared with a cluster report a negative distance,
// which must not count as a match: the reference samples belong to one model's
// vector space, so a vector of a different length says nothing about them.
func (c Clusters) Contains(other Embedding) bool {
	for _, cluster := range c {
		if cluster.Disabled {
			continue
		} else if d := cluster.Embedding.Dist(other); d >= 0 && d < cluster.Radius {
			return true
		}
	}

	return false
}

// Dist returns the minimum distance between the embedding and the enabled
// clusters. The result is -1 when no enabled clusters are present, matching
// Embedding.Dist semantics for unsupported comparisons.
func (c Clusters) Dist(other Embedding) (dist float64) {
	dist = -1

	for _, cluster := range c {
		if cluster.Disabled {
			continue
		} else if d := cluster.Embedding.Dist(other); d < 0 {
			continue
		} else if d < dist || dist < 0 {
			dist = d
		}
	}

	return dist
}
