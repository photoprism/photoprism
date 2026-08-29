package alg

import "math"

// reachHeap is a min-heap of point indices keyed by reachability, with the position of every
// queued point tracked so that an improved reachability updates it in place.
//
// In place rather than by re-inserting: OPTICS improves a point's reachability once per expansion
// that reaches it, so lazy duplicates would grow the queue toward n^2 entries on a dense set.
type reachHeap struct {
	items []int
	pos   []int
	prio  []float64
}

// newReachHeap returns a heap able to hold the given number of points.
func newReachHeap(n int) *reachHeap {
	h := &reachHeap{
		items: make([]int, 0, n),
		pos:   make([]int, n),
		prio:  make([]float64, n),
	}

	h.reset()

	return h
}

// reset empties the heap, keeping the storage for the next expansion.
func (h *reachHeap) reset() {
	h.items = h.items[:0]

	for i := range h.pos {
		h.pos[i] = -1
		h.prio[i] = math.Inf(1)
	}
}

// Len returns the number of queued points.
func (h *reachHeap) Len() int {
	return len(h.items)
}

// push queues a point, or moves an already queued one up when its reachability has improved.
func (h *reachHeap) push(p int, prio float64) {
	if i := h.pos[p]; i >= 0 {
		if prio >= h.prio[p] {
			return
		}

		h.prio[p] = prio
		h.up(i)

		return
	}

	h.prio[p] = prio
	h.items = append(h.items, p)
	h.pos[p] = len(h.items) - 1
	h.up(len(h.items) - 1)
}

// pop removes and returns the point with the shortest reachability, or -1 when the heap is empty.
func (h *reachHeap) pop() int {
	if len(h.items) == 0 {
		return -1
	}

	top := h.items[0]
	last := len(h.items) - 1

	h.swap(0, last)
	h.items = h.items[:last]
	h.pos[top] = -1

	if last > 0 {
		h.down(0)
	}

	return top
}

// swap exchanges two heap positions, keeping the position index in step.
func (h *reachHeap) swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.pos[h.items[i]] = i
	h.pos[h.items[j]] = j
}

// less reports whether the point at position i sorts before the one at j, breaking ties by point
// index so that an ordering does not depend on the order equal distances were queued in.
func (h *reachHeap) less(i, j int) bool {
	a, b := h.items[i], h.items[j]

	if h.prio[a] != h.prio[b] {
		return h.prio[a] < h.prio[b]
	}

	return a < b
}

// up moves a point toward the root until its parent sorts before it.
func (h *reachHeap) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2

		if !h.less(i, parent) {
			return
		}

		h.swap(i, parent)
		i = parent
	}
}

// down moves a point toward the leaves until both children sort after it.
func (h *reachHeap) down(i int) {
	for {
		left := 2*i + 1

		if left >= len(h.items) {
			return
		}

		smallest := left

		if right := left + 1; right < len(h.items) && h.less(right, left) {
			smallest = right
		}

		if !h.less(smallest, i) {
			return
		}

		h.swap(i, smallest)
		i = smallest
	}
}
