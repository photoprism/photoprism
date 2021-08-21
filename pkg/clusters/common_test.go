package clusters

import (
	"testing"
)

func TestQueueEmptyWhenCreated(t *testing.T) {
	queue := newPriorityQueue(0)
	if queue.NotEmpty() {
		t.Error("Newly created queue is not empty")
	}
}

func TestQueueNowEmptyAfterAdd(t *testing.T) {
	queue := newPriorityQueue(1)

	queue.Push(&pItem{
		v: 0,
		p: 0.5,
	})

	if !queue.NotEmpty() {
		t.Error("Queue is empty after a single add")
	}
}

func TestQueueReturnsInPriorityOrder(t *testing.T) {
	queue := newPriorityQueue(2)

	var (
		itemOne = &pItem{
			v: 0,
			p: 0.5,
		}
		itemTwo = &pItem{
			v: 1,
			p: 0.6,
		}
	)

	queue.Push(itemTwo)
	queue.Push(itemOne)

	if queue.Pop().(*pItem) != itemOne {
		t.Error("Queue is should return itemOne first")
	}

	if queue.Pop().(*pItem) != itemTwo {
		t.Error("Queue is should return itemTwo next")
	}

	if queue.NotEmpty() {
		t.Error("Queue is not empty")
	}
}

func TestQueueReturnsInPriorityOrderAfterUpdate(t *testing.T) {
	queue := newPriorityQueue(2)

	var (
		itemOne = &pItem{
			v: 0,
			p: 0.5,
		}
		itemTwo = &pItem{
			v: 1,
			p: 0.6,
		}
	)

	queue.Push(itemTwo)
	queue.Push(itemOne)

	queue.Update(itemTwo, 1, 0.4)

	if queue.Pop().(*pItem) != itemTwo {
		t.Error("Queue is should return itemTwo first")
	}

	if queue.Pop().(*pItem) != itemOne {
		t.Error("Queue is should return itemOne next")
	}

	if queue.NotEmpty() {
		t.Error("Queue is not empty")
	}
}

func TestBounds(t *testing.T) {
	var (
		f = "data/test.csv"
		i = CsvImporter()
		l = 3
	)

	d, e := i.Import(f, 0, 2)
	if e != nil {
		t.Errorf("Error importing data: %s\n", e.Error())
	}

	bounds := bounds(d)

	if len(bounds) != 3 {
		t.Errorf("Mismatched bounds array length: %d vs %d\n", len(bounds), l)
	}

	if bounds[0][0] != 0.1 || bounds[0][1] != 0.7 {
		t.Error("Invalid bounds for feature #0")
	}

	if bounds[1][0] != 0.2 || bounds[1][1] != 0.8 {
		t.Error("Invalid bounds for feature #1")
	}

	if bounds[2][0] != 0.3 || bounds[2][1] != 0.9 {
		t.Error("Invalid bounds for feature #2")
	}
}

func TestUniform(t *testing.T) {
	var (
		l = 100
		d = &[2]float64{
			0,
			10,
		}
	)

	for i := 0; i < l; i++ {
		u := uniform(d)
		if u < 0 || u > 10 {
			t.Error("Unformly distributed variable out of bounds")
		}
	}
}
