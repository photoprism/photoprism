package alg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReachHeap(t *testing.T) {
	t.Run("PopsInAscendingOrder", func(t *testing.T) {
		h := newReachHeap(5)
		for p, prio := range map[int]float64{0: 0.5, 1: 0.1, 2: 0.9, 3: 0.3} {
			h.push(p, prio)
		}
		assert.Equal(t, 4, h.Len())
		assert.Equal(t, []int{1, 3, 0, 2}, []int{h.pop(), h.pop(), h.pop(), h.pop()})
		assert.Equal(t, 0, h.Len())
	})
	t.Run("EmptyPopReportsNoPoint", func(t *testing.T) {
		assert.Equal(t, -1, newReachHeap(3).pop())
	})
	t.Run("ImprovedPriorityMovesUp", func(t *testing.T) {
		// The reason for tracking positions: re-inserting instead would leave the stale entry
		// behind, and OPTICS improves a reachability once per expansion that reaches the point.
		h := newReachHeap(3)
		h.push(0, 0.1)
		h.push(1, 0.9)
		h.push(1, 0.05)
		assert.Equal(t, 2, h.Len(), "an improved point may not be queued twice")
		assert.Equal(t, 1, h.pop())
		assert.Equal(t, 0, h.pop())
	})
	t.Run("WorsePriorityIgnored", func(t *testing.T) {
		h := newReachHeap(3)
		h.push(0, 0.1)
		h.push(0, 0.8)
		assert.Equal(t, 1, h.Len())
		assert.Equal(t, 0, h.pop())
	})
	t.Run("TiesBreakByPointIndex", func(t *testing.T) {
		// Equal reachabilities are common, since several points reached from one predecessor all
		// take its core distance. Ordering them by index keeps a pass reproducible.
		h := newReachHeap(4)
		for _, p := range []int{3, 1, 2, 0} {
			h.push(p, 0.25)
		}
		assert.Equal(t, []int{0, 1, 2, 3}, []int{h.pop(), h.pop(), h.pop(), h.pop()})
	})
	t.Run("ResetKeepsTheHeapUsable", func(t *testing.T) {
		h := newReachHeap(3)
		h.push(0, 0.1)
		h.reset()
		assert.Equal(t, 0, h.Len())
		h.push(2, 0.7)
		assert.Equal(t, 2, h.pop())
	})
	t.Run("PopAfterResetSeesTheNewPriority", func(t *testing.T) {
		// reset has to clear the recorded priorities as well, or a point queued again in the next
		// expansion would be compared against the value it carried in the previous one.
		h := newReachHeap(2)
		h.push(0, 0.1)
		h.pop()
		h.reset()
		h.push(0, 0.9)
		assert.Equal(t, 1, h.Len())
		assert.Equal(t, 0, h.pop())
	})
}

func TestUnionFind(t *testing.T) {
	t.Run("JoinsComponents", func(t *testing.T) {
		u := newUnionFind(5)
		assert.NotEqual(t, u.find(0), u.find(1))
		u.union(0, 1)
		u.union(1, 2)
		assert.Equal(t, u.find(0), u.find(2))
		assert.NotEqual(t, u.find(0), u.find(3))
	})
	t.Run("UnionOfOneComponentIsStable", func(t *testing.T) {
		u := newUnionFind(3)
		root := u.union(0, 1)
		assert.Equal(t, root, u.union(0, 1))
	})
}

func TestAncestry(t *testing.T) {
	t.Run("WalksToTheNearestKeptNode", func(t *testing.T) {
		a := ancestry{}
		a.attach(12, 11)
		a.attach(11, 10)
		assert.Equal(t, 10, a.root(12))
		assert.Equal(t, 10, a.root(10))
	})
	t.Run("UnattachedNodeIsItsOwnRoot", func(t *testing.T) {
		assert.Equal(t, 7, ancestry{}.root(7))
	})
	t.Run("SelfReferenceTerminates", func(t *testing.T) {
		a := ancestry{5: 5}
		assert.Equal(t, 5, a.root(5))
	})
}
