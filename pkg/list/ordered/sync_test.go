package ordered

import (
	"fmt"
	"math/rand" //nolint:gosec // pseudo-random is sufficient for concurrency tests
	"sync"
	"testing"
)

func TestRaceCondition(t *testing.T) {
	m := NewSyncMap[int, int]()
	wg := &sync.WaitGroup{}
	//nolint:gosec // pseudo-random is sufficient for race testing
	randInt := func() int {
		return rand.Intn(100)
	}

	var asyncGet = func() {
		wg.Add(1)
		go func() {
			key := randInt()
			m.Get(key)
			wg.Done()
		}()
	}

	var asyncSet = func() {
		wg.Add(1)
		go func() {
			key := randInt()
			value := randInt()
			m.Set(key, value)
			wg.Done()
		}()
	}

	var asyncDelete = func() {
		wg.Add(1)
		go func() {
			key := randInt()
			m.Delete(key)
			wg.Done()
		}()
	}

	var asyncHas = func() {
		wg.Add(1)
		go func() {
			key := randInt()
			m.Has(key)
			wg.Done()
		}()
	}

	var asyncReplaceKEy = func() {
		wg.Add(1)
		go func() {
			key := randInt()
			newKey := randInt()
			m.ReplaceKey(key, newKey)
			wg.Done()
		}()
	}

	var asyncGetOrDefault = func() {
		wg.Add(1)
		go func() {
			key := randInt()
			def := randInt()
			m.GetOrDefault(key, def)
			wg.Done()
		}()
	}

	var asyncLen = func() {
		wg.Add(1)
		go func() {
			m.Len()
			wg.Done()
		}()
	}

	var asyncCopy = func() {
		wg.Add(1)
		go func() {
			m.Copy()
			wg.Done()
		}()
	}

	for i := 0; i < 10000; i++ {
		asyncSet()
		asyncGet()
		asyncDelete()
		asyncHas()
		asyncLen()
		asyncReplaceKEy()
		asyncGetOrDefault()
		asyncCopy()
	}

	wg.Wait()
	fmt.Println("TestRaceCondition completed")
	fmt.Printf("SyncMap eventually has %v elements\n", m.Len())
}
