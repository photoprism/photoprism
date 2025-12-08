package vision

import (
	"testing"
)

func TestPriorityFromTopicality(t *testing.T) {
	cases := []struct {
		top float32
		exp int
	}{
		{0.95, 5},
		{0.90, 4},
		{0.80, 3},
		{0.65, 2},
		{0.50, 1},
		{0.40, 1},
		{0.35, -1},
		{0.05, -2},
	}

	for _, tc := range cases {
		if got := PriorityFromTopicality(tc.top); got != tc.exp {
			t.Fatalf("topicality %v expected priority %d, got %d", tc.top, tc.exp, got)
		}
	}
}
