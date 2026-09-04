package alg

// unionFind tracks which points have been joined into the same component, which is how the
// single-linkage hierarchy is built from spanning tree edges taken in order of distance.
type unionFind struct {
	parent []int
	rank   []int
}

// newUnionFind returns a structure over n elements, each in its own component.
func newUnionFind(n int) *unionFind {
	u := &unionFind{
		parent: make([]int, n),
		rank:   make([]int, n),
	}

	for i := range n {
		u.parent[i] = i
	}

	return u
}

// find returns the representative of an element's component, flattening the path it walked.
func (u *unionFind) find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}

	return x
}

// union merges two components and returns the representative of the result.
func (u *unionFind) union(a, b int) int {
	ra, rb := u.find(a), u.find(b)

	if ra == rb {
		return ra
	}

	if u.rank[ra] < u.rank[rb] {
		ra, rb = rb, ra
	}

	u.parent[rb] = ra

	if u.rank[ra] == u.rank[rb] {
		u.rank[ra]++
	}

	return ra
}

// ancestry maps a node to the nearest ancestor that was kept, by folding every node that was not
// into its parent. Unlike unionFind it keeps the direction of the tree, which is what makes the
// walk stop at the first surviving cluster rather than at an arbitrary representative.
type ancestry map[int]int

// attach records that a node was folded into its parent.
func (a ancestry) attach(node, parent int) {
	a[node] = parent
}

// root returns the nearest ancestor of a node that was not folded away, or the node itself.
func (a ancestry) root(node int) int {
	for {
		parent, ok := a[node]

		if !ok || parent == node {
			return node
		}

		node = parent
	}
}
