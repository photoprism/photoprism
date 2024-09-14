package clusters

import (
	"container/heap"
	"math/rand/v2"
	"sync"
)

// struct denoting start and end indices of database portion to be scanned for nearest neighbours by workers in DBSCAN and OPTICS
type rangeJob struct {
	a, b int
}

// priority queue
type pItem struct {
	v int
	p float64
	i int
}

type priorityQueue []*pItem

func newPriorityQueue(size int) priorityQueue {
	q := make(priorityQueue, 0, size)
	heap.Init(&q)

	return q
}

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].p > pq[j].p
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].i = i
	pq[j].i = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*pItem)
	item.i = n
	*pq = append(*pq, item)
	heap.Fix(pq, item.i)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.i = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) NotEmpty() bool {
	return len(*pq) > 0
}

func (pq *priorityQueue) Update(item *pItem, value int, priority float64) {
	item.v = value
	item.p = priority
	heap.Fix(pq, item.i)
}

func bounds(data [][]float64) []*[2]float64 {
	var (
		wg sync.WaitGroup

		l = len(data[0])
		r = make([]*[2]float64, l)
	)

	for i := 0; i < l; i++ {
		r[i] = &[2]float64{
			data[0][i],
			data[0][i],
		}
	}

	wg.Add(l)

	for i := 0; i < l; i++ {
		go func(n int) {
			defer wg.Done()

			for j := 0; j < len(data); j++ {
				if data[j][n] < r[n][0] {
					r[n][0] = data[j][n]
				} else if data[j][n] > r[n][1] {
					r[n][1] = data[j][n]
				}
			}
		}(i)
	}

	wg.Wait()

	return r
}

func uniform(data *[2]float64) float64 {
	return rand.Float64()*(data[1]-data[0]) + data[0]
}
