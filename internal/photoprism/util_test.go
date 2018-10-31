package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRandomInt(t *testing.T) {
	min := 5
	max := 50

	for i := 0; i < 10; i++ {
		result := getRandomInt(min, max)

		if result > max {
			t.Errorf("Random result must not be bigger than %d", max)
		} else if result < min {
			t.Errorf("Random result must not be smaller than %d", min)
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"zzz", "AAA", "ZZZ", "aaa", "foo", "1", "", "zzz", "AAA", "ZZZ", "aaa"}

	output := uniqueStrings(input)

	expected := []string{"foo", "1", "zzz", "AAA", "ZZZ", "aaa"}

	assert.ElementsMatch(t, expected, output)
}
