package ordered_test

//revive:disable:var-naming // benchmark helpers follow Go benchmark naming with underscores

import (
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/list/ordered"
)

func TestNewMap(t *testing.T) {
	m := ordered.NewMap[int, string]()
	assert.IsType(t, &ordered.Map[int, string]{}, m)
}

func TestGet(t *testing.T) {
	t.Run("ReturnsNotOKIfStringKeyDoesntExist", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		_, ok := m.Get("foo")
		assert.False(t, ok)
	})

	t.Run("ReturnsNotOKIfNonStringKeyDoesntExist", func(t *testing.T) {
		m := ordered.NewMap[int, string]()
		_, ok := m.Get(123)
		assert.False(t, ok)
	})

	t.Run("ReturnsOKIfKeyExists", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		_, ok := m.Get("foo")
		assert.True(t, ok)
	})

	t.Run("ReturnsValueForKey", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		value, _ := m.Get("foo")
		assert.Equal(t, "bar", value)
	})

	t.Run("ReturnsDynamicValueForKey", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "baz")
		value, _ := m.Get("foo")
		assert.Equal(t, "baz", value)
	})

	t.Run("KeyDoesntExistOnNonEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "baz")
		_, ok := m.Get("bar")
		assert.False(t, ok)
	})

	t.Run("ValueForKeyDoesntExistOnNonEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "baz")
		value, _ := m.Get("bar")
		assert.Empty(t, value)
	})
}

func TestSet(t *testing.T) {
	t.Run("ReturnsTrueIfStringKeyIsNew", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		ok := m.Set("foo", "bar")
		assert.True(t, ok)
	})

	t.Run("ReturnsTrueIfNonStringKeyIsNew", func(t *testing.T) {
		m := ordered.NewMap[int, string]()
		ok := m.Set(123, "bar")
		assert.True(t, ok)
	})

	t.Run("ValueCanBeNonString", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		ok := m.Set(123, true)
		assert.True(t, ok)
	})

	t.Run("ReturnsFalseIfKeyIsNotNew", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		ok := m.Set("foo", "bar")
		assert.False(t, ok)
	})

	t.Run("SetThreeDifferentKeys", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		m.Set("baz", "qux")
		ok := m.Set("quux", "corge")
		assert.True(t, ok)
	})
}

func TestReplaceKey(t *testing.T) {
	t.Run("ReturnsFalseIfOriginalKeyDoesntExist", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		assert.False(t, m.ReplaceKey("foo", "bar"))
	})

	t.Run("ReturnsFalseIfNewKeyAlreadyExists", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		m.Set("baz", "qux")
		assert.False(t, m.ReplaceKey("foo", "baz"))
		assert.Equal(t, []string{"foo", "baz"}, slices.Collect(m.Keys()))
	})

	t.Run("ReturnsTrueIfOnlyOriginalKeyExists", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		assert.True(t, m.ReplaceKey("foo", "baz"))

		// Now validate the "replacement" was a success.
		el := m.GetElement("baz")
		require.NotNil(t, el)
		assert.Equal(t, "bar", el.Value)
		assert.Equal(t, "baz", el.Key)

		v, ok := m.Get("baz")
		assert.True(t, ok)
		assert.Equal(t, "bar", v)
		assert.Equal(t, []string{"baz"}, slices.Collect(m.Keys()))
		assert.Equal(t, 1, m.Len())

		_, ok = m.Get("foo") // original key
		assert.False(t, ok)
	})

	t.Run("KeyMaintainsOrderWhenReplaced", func(t *testing.T) {
		count := 100
		// Build a larger map to help validate that the order is not coincidental.
		m := ordered.NewMap[int, int]()
		for i := 0; i < count; i++ {
			m.Set(i, i)
		}
		// Rename the middle 50-60 elements to 100+ current
		for i := 50; i < 60; i++ {
			assert.True(t, m.ReplaceKey(i, i+100))
		}

		// ensure length is maintained
		assert.Equal(t, count, m.Len())

		// Validate the order is maintained.
		for i, key := range slices.Collect(m.Keys()) {
			if i >= 50 && i < 60 {
				assert.Equal(t, i+100, key)
			} else {
				assert.Equal(t, i, key)
			}
		}
	})
}

func TestLen(t *testing.T) {
	t.Run("EmptyMapIsZeroLen", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		assert.Equal(t, 0, m.Len())
	})

	t.Run("SingleElementIsLenOne", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		m.Set(123, true)
		assert.Equal(t, 1, m.Len())
	})

	t.Run("ThreeElements", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		m.Set(1, true)
		m.Set(2, true)
		m.Set(3, true)
		assert.Equal(t, 3, m.Len())
	})
}

func TestKeys(t *testing.T) {
	t.Run("EmptyMap", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		assert.Empty(t, slices.Collect(m.Keys()))
	})

	t.Run("OneElement", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		m.Set(1, true)
		assert.Equal(t, []int{1}, slices.Collect(m.Keys()))
	})

	t.Run("RetainsOrder", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		for i := 1; i < 10; i++ {
			m.Set(i, true)
		}
		assert.Equal(t,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			slices.Collect(m.Keys()))
	})

	t.Run("ReplacingKeyDoesntChangeOrder", func(t *testing.T) {
		m := ordered.NewMap[string, bool]()
		m.Set("foo", true)
		m.Set("bar", true)
		m.Set("foo", false)
		assert.Equal(t,
			[]string{"foo", "bar"},
			slices.Collect(m.Keys()))
	})

	t.Run("KeysAfterDelete", func(t *testing.T) {
		m := ordered.NewMap[string, bool]()
		m.Set("foo", true)
		m.Set("bar", true)
		m.Delete("foo")
		assert.Equal(t, []string{"bar"}, slices.Collect(m.Keys()))
	})
}

func TestDelete(t *testing.T) {
	t.Run("KeyDoesntExistReturnsFalse", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		assert.False(t, m.Delete("foo"))
	})

	t.Run("KeyDoesExist", func(t *testing.T) {
		m := ordered.NewMap[string, any]()
		m.Set("foo", nil)
		assert.True(t, m.Delete("foo"))
	})

	t.Run("KeyNoLongerExists", func(t *testing.T) {
		m := ordered.NewMap[string, any]()
		m.Set("foo", nil)
		m.Delete("foo")
		_, exists := m.Get("foo")
		assert.False(t, exists)
	})

	t.Run("KeyDeleteIsIsolated", func(t *testing.T) {
		m := ordered.NewMap[string, any]()
		m.Set("foo", nil)
		m.Set("bar", nil)
		m.Delete("foo")
		_, exists := m.Get("bar")
		assert.True(t, exists)
	})
}

func TestMap_Front(t *testing.T) {
	t.Run("NilOnEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		assert.Nil(t, m.Front())
	})

	t.Run("NilOnEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		m.Set(1, true)
		assert.NotNil(t, m.Front())
	})
}

func TestMap_Back(t *testing.T) {
	t.Run("NilOnEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		assert.Nil(t, m.Back())
	})

	t.Run("NilOnEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		m.Set(1, true)
		assert.NotNil(t, m.Back())
	})
}

func TestMap_Copy(t *testing.T) {
	t.Run("ReturnsEqualButNotSame", func(t *testing.T) {
		key, value := 1, "a value"
		m := ordered.NewMap[int, string]()
		m.Set(key, value)

		m2 := m.Copy()
		m2.Set(key, "a different value")

		assert.Equal(t, m.Len(), m2.Len(), "not all elements are copied")
		assert.Equal(t, value, m.GetElement(key).Value)
	})
}

func TestGetElement(t *testing.T) {
	t.Run("ReturnsElementForKey", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")

		var results []any
		element := m.GetElement("foo")
		if element != nil {
			results = append(results, element.Key, element.Value)
		}

		assert.Equal(t, []any{"foo", "bar"}, results)
	})

	t.Run("ElementForKeyDoesntExistOnNonEmptyMap", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "baz")
		element := m.GetElement("bar")
		assert.Nil(t, element)
	})
}

func TestSetAndGet(t *testing.T) {
	t.Run("FourBoolElements", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		expected := map[int]bool{1: true, 3: false, 5: false, 4: true}
		for k, v := range expected {
			m.Set(k, v)
		}
		for k, v := range expected {
			w, ok := m.Get(k)
			assert.True(t, ok)
			assert.Equal(t, v, w)
		}
	})
}

func TestIterations(t *testing.T) {
	type Element struct {
		Key   int
		Value bool
	}
	t.Run("FourBoolElements", func(t *testing.T) {
		m := ordered.NewMap[int, bool]()
		expected := []Element{{5, true}, {3, false}, {1, false}, {4, true}}
		for _, v := range expected {
			m.Set(v.Key, v.Value)
		}
		element := m.Front()
		for i := 0; i < len(expected); i++ {
			assert.NotNil(t, element)
			assert.Equal(t, expected[i].Key, element.Key)
			assert.Equal(t, expected[i].Value, element.Value)
			element = element.Next()
		}
		assert.Nil(t, element)
	})
}

func TestIterators(t *testing.T) {
	type Element struct {
		Key   int
		Value bool
	}
	m := ordered.NewMap[int, bool]()
	expected := []Element{{5, true}, {3, false}, {1, false}, {4, true}}
	for _, v := range expected {
		m.Set(v.Key, v.Value)
	}

	t.Run("Iterator", func(t *testing.T) {
		i := 0
		for key, value := range m.AllFromFront() {
			assert.Equal(t, expected[i].Key, key)
			assert.Equal(t, expected[i].Value, value)
			i++
		}
	})

	t.Run("ReverseIterator", func(t *testing.T) {
		i := len(expected) - 1
		for key, value := range m.AllFromBack() {
			assert.Equal(t, expected[i].Key, key)
			assert.Equal(t, expected[i].Value, value)
			i--
		}
	})
}

func TestMap_Has(t *testing.T) {
	t.Run("ReturnsFalseIfKeyDoesNotExist", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		assert.False(t, m.Has("foo"))
	})

	t.Run("ReturnsTrueIfKeyExists", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		assert.True(t, m.Has("foo"))
	})

	t.Run("KeyDoesNotExistAfterDelete", func(t *testing.T) {
		m := ordered.NewMap[string, string]()
		m.Set("foo", "bar")
		m.Delete("foo")
		assert.False(t, m.Has("foo"))
	})
}

func benchmarkMap_Set(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := make(map[int]bool)
		for i := 0; i < b.N*multiplier; i++ {
			m[i] = true
		}
	}
}

func BenchmarkMap_Set(b *testing.B) {
	benchmarkMap_Set(1)(b)
}

func benchmarkOrderedMap_Set(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := ordered.NewMap[int, bool]()
		for i := 0; i < b.N*multiplier; i++ {
			m.Set(i, true)
		}
	}
}

func BenchmarkOrderedMap_Set(b *testing.B) {
	benchmarkOrderedMap_Set(1)(b)
}

func benchmarkMap_Get(multiplier int) func(b *testing.B) {
	m := make(map[int]bool)
	for i := 0; i < 1000*multiplier; i++ {
		m[i] = true
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = m[i%1000*multiplier]
		}
	}
}

func BenchmarkMap_Get(b *testing.B) {
	benchmarkMap_Get(1)(b)
}

func benchmarkOrderedMap_Get(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.Get(i % 1000 * multiplier)
		}
	}
}

func BenchmarkOrderedMap_Get(b *testing.B) {
	benchmarkOrderedMap_Get(1)(b)
}

func benchmarkOrderedMap_GetElement(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.GetElement(i % 1000 * multiplier)
		}
	}
}

func BenchmarkOrderedMap_GetElement(b *testing.B) {
	benchmarkOrderedMap_GetElement(1)(b)
}

var tempInt int

func benchmarkOrderedMap_Len(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		var temp int
		for i := 0; i < b.N; i++ {
			temp = m.Len()
		}

		// prevent compiler from optimizing Len away.
		tempInt = temp
	}
}

func BenchmarkOrderedMap_Len(b *testing.B) {
	benchmarkOrderedMap_Len(1)(b)
}

func benchmarkMap_Delete(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := make(map[int]bool)
		for i := 0; i < b.N*multiplier; i++ {
			m[i] = true
		}

		for i := 0; i < b.N; i++ {
			delete(m, i)
		}
	}
}

func BenchmarkMap_Delete(b *testing.B) {
	benchmarkMap_Delete(1)(b)
}

func benchmarkOrderedMap_Delete(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := ordered.NewMap[int, bool]()
		for i := 0; i < b.N*multiplier; i++ {
			m.Set(i, true)
		}

		for i := 0; i < b.N; i++ {
			m.Delete(i)
		}
	}
}

func BenchmarkOrderedMap_Delete(b *testing.B) {
	benchmarkOrderedMap_Delete(1)(b)
}

func benchmarkMap_Iterate(multiplier int) func(b *testing.B) {
	m := make(map[int]bool)
	for i := 0; i < 1000*multiplier; i++ {
		m[i] = true
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, v := range m {
				nothing(v)
			}
		}
	}
}

func BenchmarkMap_Iterate(b *testing.B) {
	benchmarkMap_Iterate(1)(b)
}

func benchmarkOrderedMap_Iterate(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, key := range slices.Collect(m.Keys()) {
				_, v := m.Get(key)
				nothing(v)
			}
		}
	}
}

func BenchmarkOrderedMap_Iterate(b *testing.B) {
	benchmarkOrderedMap_Iterate(1)(b)
}

func benchmarkMap_Has(multiplier int) func(b *testing.B) {
	m := make(map[int]bool)
	for i := 0; i < 1000*multiplier; i++ {
		m[i] = true
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = m[i%1000*multiplier]
		}
	}
}

func BenchmarkMap_Has(b *testing.B) {
	benchmarkMap_Has(1)(b)
}

func benchmarkOrderedMap_Has(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(i, true)
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.Has(i % 1000 * multiplier)
		}
	}
}

func BenchmarkOrderedMap_Has(b *testing.B) {
	benchmarkOrderedMap_Has(1)(b)
}

func benchmarkMapString_Set(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := make(map[string]bool)
		a := "12345678"
		for i := 0; i < b.N*multiplier; i++ {
			m[a+strconv.Itoa(i)] = true
		}
	}
}

func BenchmarkMapString_Set(b *testing.B) {
	benchmarkMapString_Set(1)(b)
}

func benchmarkOrderedMapString_Set(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := ordered.NewMap[string, bool]()
		a := "12345678"
		for i := 0; i < b.N*multiplier; i++ {
			m.Set(a+strconv.Itoa(i), true)
		}
	}
}

func BenchmarkOrderedMapString_Set(b *testing.B) {
	benchmarkOrderedMapString_Set(1)(b)
}

func benchmarkMapString_Get(multiplier int) func(b *testing.B) {
	m := make(map[string]bool)
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m[a+strconv.Itoa(i)] = true
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = m[a+strconv.Itoa(i%1000*multiplier)]
		}
	}
}

func BenchmarkMapString_Get(b *testing.B) {
	benchmarkMapString_Get(1)(b)
}

func benchmarkOrderedMapString_Get(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.Get(a + strconv.Itoa(i%1000*multiplier))
		}
	}
}

func BenchmarkOrderedMapString_Get(b *testing.B) {
	benchmarkOrderedMapString_Get(1)(b)
}

func benchmarkOrderedMapString_GetElement(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.GetElement(a + strconv.Itoa(i%1000*multiplier))
		}
	}
}

func BenchmarkOrderedMapString_GetElement(b *testing.B) {
	benchmarkOrderedMapString_GetElement(1)(b)
}

func benchmarkMapString_Delete(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := make(map[string]bool)
		a := "12345678"
		for i := 0; i < b.N*multiplier; i++ {
			m[a+strconv.Itoa(i)] = true
		}

		for i := 0; i < b.N; i++ {
			delete(m, a+strconv.Itoa(i))
		}
	}
}

func BenchmarkMapString_Delete(b *testing.B) {
	benchmarkMapString_Delete(1)(b)
}

func benchmarkOrderedMapString_Delete(multiplier int) func(b *testing.B) {
	return func(b *testing.B) {
		m := ordered.NewMap[string, bool]()
		a := "12345678"
		for i := 0; i < b.N*multiplier; i++ {
			m.Set(a+strconv.Itoa(i), true)
		}

		for i := 0; i < b.N; i++ {
			m.Delete(a + strconv.Itoa(i))
		}
	}
}

func BenchmarkOrderedMapString_Delete(b *testing.B) {
	benchmarkOrderedMapString_Delete(1)(b)
}

func benchmarkMapString_Iterate(multiplier int) func(b *testing.B) {
	m := make(map[string]bool)
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m[a+strconv.Itoa(i)] = true
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, v := range m {
				nothing(v)
			}
		}
	}
}

func BenchmarkMapString_Iterate(b *testing.B) {
	benchmarkMapString_Iterate(1)(b)
}

func benchmarkOrderedMapString_Iterate(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, key := range slices.Collect(m.Keys()) {
				_, v := m.Get(key)
				nothing(v)
			}
		}
	}
}

func BenchmarkOrderedMapString_Iterate(b *testing.B) {
	benchmarkOrderedMapString_Iterate(1)(b)
}

func benchmarkMapString_Has(multiplier int) func(b *testing.B) {
	m := make(map[string]bool)
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m[a+strconv.Itoa(i)] = true
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = m[a+strconv.Itoa(i%1000*multiplier)]
		}
	}
}

func BenchmarkMapString_Has(b *testing.B) {
	benchmarkMapString_Has(1)(b)
}

func benchmarkOrderedMapString_Has(multiplier int) func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "12345678"
	for i := 0; i < 1000*multiplier; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m.Has(a + strconv.Itoa(i%1000*multiplier))
		}
	}
}

func BenchmarkOrderedMapString_Has(b *testing.B) {
	benchmarkOrderedMapString_Has(1)(b)
}

func nothing(v interface{}) {
	_ = v
}

func benchmarkBigMap_Set() func(b *testing.B) {
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			m := make(map[int]bool)
			for i := 0; i < 10000000; i++ {
				m[i] = true
			}
		}
	}
}

func BenchmarkBigMap_Set(b *testing.B) {
	benchmarkBigMap_Set()(b)
}

func benchmarkBigOrderedMap_Set() func(b *testing.B) {
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			m := ordered.NewMap[int, bool]()
			for i := 0; i < 10000000; i++ {
				m.Set(i, true)
			}
		}
	}
}

func BenchmarkBigOrderedMap_Set(b *testing.B) {
	benchmarkBigOrderedMap_Set()(b)
}

func benchmarkBigMapWithCapacity_Set() func(b *testing.B) {
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			m := ordered.NewMapWithCapacity[int, bool](10000000)
			for i := 0; i < 10000000; i++ {
				m.Set(i, true)
			}
		}
	}
}

func BenchmarkBigMapWithCapacity_Set(b *testing.B) {
	benchmarkBigMapWithCapacity_Set()(b)
}

func benchmarkBigMap_Get() func(b *testing.B) {
	m := make(map[int]bool)
	for i := 0; i < 10000000; i++ {
		m[i] = true
	}

	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				_ = m[i]
			}
		}
	}
}

func BenchmarkBigMap_Get(b *testing.B) {
	benchmarkBigMap_Get()(b)
}

func benchmarkBigOrderedMap_Get() func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 10000000; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				m.Get(i)
			}
		}
	}
}

func BenchmarkBigOrderedMap_Get(b *testing.B) {
	benchmarkBigOrderedMap_Get()(b)
}

func benchmarkBigOrderedMap_GetElement() func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 10000000; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				m.GetElement(i)
			}
		}
	}
}

func BenchmarkBigOrderedMap_GetElement(b *testing.B) {
	benchmarkBigOrderedMap_GetElement()(b)
}

func benchmarkBigMap_Iterate() func(b *testing.B) {
	m := make(map[int]bool)
	for i := 0; i < 10000000; i++ {
		m[i] = true
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, v := range m {
				nothing(v)
			}
		}
	}
}

func BenchmarkBigMap_Iterate(b *testing.B) {
	benchmarkBigMap_Iterate()(b)
}

func benchmarkBigOrderedMap_Iterate() func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 10000000; i++ {
		m.Set(i, true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, key := range slices.Collect(m.Keys()) {
				_, v := m.Get(key)
				nothing(v)
			}
		}
	}
}

func BenchmarkBigOrderedMap_Iterate(b *testing.B) {
	benchmarkBigOrderedMap_Iterate()(b)
}

func benchmarkBigMapString_Set() func(b *testing.B) {
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			m := make(map[string]bool)
			a := "1234567"
			for i := 0; i < 10000000; i++ {
				m[a+strconv.Itoa(i)] = true
			}
		}
	}
}

func benchmarkBigMap_Has() func(b *testing.B) {
	m := make(map[int]bool)
	for i := 0; i < 10000000; i++ {
		m[i] = true
	}
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				_ = m[i]
			}
		}
	}
}

func BenchmarkBigMap_Has(b *testing.B) {
	benchmarkBigMap_Has()(b)
}

func benchmarkBigOrderedMap_Has() func(b *testing.B) {
	m := ordered.NewMap[int, bool]()
	for i := 0; i < 10000000; i++ {
		m.Set(i, true)
	}
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				m.Has(i)
			}
		}
	}
}

func BenchmarkBigOrderedMap_Has(b *testing.B) {
	benchmarkBigOrderedMap_Has()(b)
}

func BenchmarkBigMapString_Set(b *testing.B) {
	benchmarkBigMapString_Set()(b)
}

func benchmarkBigOrderedMapString_Set() func(b *testing.B) {
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			m := ordered.NewMap[string, bool]()
			a := "1234567"
			for i := 0; i < 10000000; i++ {
				m.Set(a+strconv.Itoa(i), true)
			}
		}
	}
}

func BenchmarkBigOrderedMapString_Set(b *testing.B) {
	benchmarkBigOrderedMapString_Set()(b)
}

func benchmarkBigMapString_Get() func(b *testing.B) {
	m := make(map[string]bool)
	a := "1234567"
	for i := 0; i < 10000000; i++ {
		m[a+strconv.Itoa(i)] = true
	}

	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				_ = m[a+strconv.Itoa(i)]
			}
		}
	}
}

func BenchmarkBigMapString_Get(b *testing.B) {
	benchmarkBigMapString_Get()(b)
}

func benchmarkBigOrderedMapString_Get() func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "1234567"
	for i := 0; i < 10000000; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}

	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				m.Get(a + strconv.Itoa(i))
			}
		}
	}
}

func BenchmarkBigOrderedMapString_Get(b *testing.B) {
	benchmarkBigOrderedMapString_Get()(b)
}

func benchmarkBigOrderedMapString_GetElement() func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "1234567"
	for i := 0; i < 10000000; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}

	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				m.GetElement(a + strconv.Itoa(i))
			}
		}
	}
}

func BenchmarkBigOrderedMapString_GetElement(b *testing.B) {
	benchmarkBigOrderedMapString_GetElement()(b)
}

func benchmarkBigMapString_Iterate() func(b *testing.B) {
	m := make(map[string]bool)
	a := "12345678"
	for i := 0; i < 10000000; i++ {
		m[a+strconv.Itoa(i)] = true
	}
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, v := range m {
				nothing(v)
			}
		}
	}
}

func BenchmarkBigMapString_Iterate(b *testing.B) {
	benchmarkBigMapString_Iterate()(b)
}

func benchmarkBigOrderedMapString_Iterate() func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "12345678"
	for i := 0; i < 10000000; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}

	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, key := range slices.Collect(m.Keys()) {
				_, v := m.Get(key)
				nothing(v)
			}
		}
	}
}

func BenchmarkBigOrderedMapString_Iterate(b *testing.B) {
	benchmarkBigOrderedMapString_Iterate()(b)
}

func benchmarkBigMapString_Has() func(b *testing.B) {
	m := make(map[string]bool)
	a := "12345678"
	for i := 0; i < 10000000; i++ {
		m[a+strconv.Itoa(i)] = true
	}
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				_ = m[a+strconv.Itoa(i)]
			}
		}
	}
}

func BenchmarkBigMapString_Has(b *testing.B) {
	benchmarkBigMapString_Has()(b)
}

func benchmarkBigOrderedMapString_Has() func(b *testing.B) {
	m := ordered.NewMap[string, bool]()
	a := "12345678"
	for i := 0; i < 10000000; i++ {
		m.Set(a+strconv.Itoa(i), true)
	}
	return func(b *testing.B) {
		for j := 0; j < b.N; j++ {
			for i := 0; i < 10000000; i++ {
				m.Has(a + strconv.Itoa(i))
			}
		}
	}
}

func BenchmarkBigOrderedMapString_Has(b *testing.B) {
	benchmarkBigOrderedMapString_Has()(b)
}

func BenchmarkAll(b *testing.B) {
	b.Run("BenchmarkOrderedMap_Set", BenchmarkOrderedMap_Set)
	b.Run("BenchmarkMap_Set", BenchmarkMap_Set)
	b.Run("BenchmarkOrderedMap_Get", BenchmarkOrderedMap_Get)
	b.Run("BenchmarkMap_Get", BenchmarkMap_Get)
	b.Run("BenchmarkOrderedMap_GetElement", BenchmarkOrderedMap_GetElement)
	b.Run("BenchmarkOrderedMap_Delete", BenchmarkOrderedMap_Delete)
	b.Run("BenchmarkMap_Delete", BenchmarkMap_Delete)
	b.Run("BenchmarkOrderedMap_Iterate", BenchmarkOrderedMap_Iterate)
	b.Run("BenchmarkMap_Iterate", BenchmarkMap_Iterate)
	b.Run("BenchmarkOrderedMap_Has", BenchmarkOrderedMap_Has)
	b.Run("BenchmarkMap_Has", BenchmarkMap_Has)

	b.Run("BenchmarkBigMap_Set", BenchmarkBigMap_Set)
	b.Run("BenchmarkBigOrderedMap_Set", BenchmarkBigOrderedMap_Set)
	b.Run("BenchmarkBigMap_Get", BenchmarkBigMap_Get)
	b.Run("BenchmarkBigOrderedMap_Get", BenchmarkBigOrderedMap_Get)
	b.Run("BenchmarkBigOrderedMap_GetElement",
		BenchmarkBigOrderedMap_GetElement)
	b.Run("BenchmarkBigOrderedMap_Iterate", BenchmarkBigOrderedMap_Iterate)
	b.Run("BenchmarkBigMap_Iterate", BenchmarkBigMap_Iterate)
	b.Run("BenchmarkBigMap_Has", BenchmarkBigMap_Has)
	b.Run("BenchmarkBigOrderedMap_Has", BenchmarkBigOrderedMap_Has)

	b.Run("BenchmarkOrderedMapString_Set", BenchmarkOrderedMapString_Set)
	b.Run("BenchmarkMapString_Set", BenchmarkMapString_Set)
	b.Run("BenchmarkOrderedMapString_Get", BenchmarkOrderedMapString_Get)
	b.Run("BenchmarkMapString_Get", BenchmarkMapString_Get)
	b.Run("BenchmarkOrderedMapString_GetElement",
		BenchmarkOrderedMapString_GetElement)
	b.Run("BenchmarkOrderedMapString_Delete", BenchmarkOrderedMapString_Delete)
	b.Run("BenchmarkMapString_Delete", BenchmarkMapString_Delete)
	b.Run("BenchmarkOrderedMapString_Iterate",
		BenchmarkOrderedMapString_Iterate)
	b.Run("BenchmarkMapString_Iterate", BenchmarkMapString_Iterate)
	b.Run("BenchmarkMapString_Has", BenchmarkMapString_Has)
	b.Run("BenchmarkOrderedMapString_Has", BenchmarkOrderedMapString_Has)

	b.Run("BenchmarkBigMapString_Set", BenchmarkBigMapString_Set)
	b.Run("BenchmarkBigOrderedMapString_Set", BenchmarkBigOrderedMapString_Set)
	b.Run("BenchmarkBigMapString_Get", BenchmarkBigMapString_Get)
	b.Run("BenchmarkBigOrderedMapString_Get", BenchmarkBigOrderedMapString_Get)
	b.Run("BenchmarkBigOrderedMapString_GetElement",
		BenchmarkBigOrderedMapString_GetElement)
	b.Run("BenchmarkBigOrderedMapString_Iterate",
		BenchmarkBigOrderedMapString_Iterate)
	b.Run("BenchmarkBigMapString_Iterate", BenchmarkBigMapString_Iterate)
	b.Run("BenchmarkBigMapString_Has", BenchmarkBigMapString_Has)
	b.Run("BenchmarkBigOrderedMapString_Has", BenchmarkBigOrderedMapString_Has)
}
