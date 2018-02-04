package photoprism

import "testing"

func TestGetRandomInt(t *testing.T) {
	min := 5
	max := 50

	for i := 0; i < 10; i++ {
		result := GetRandomInt(min, max)

		if result > max {
			t.Errorf("Random result must not be bigger than %d", max)
		} else if result < min {
			t.Errorf("Random result must not be smaller than %d", min)
		}
	}
}
