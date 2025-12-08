package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	rows := [][]string{
		{"b", "z"},
		{"a", "b"},
		{"a", "a"}, // tie on col 0 broken by col 1
	}
	Sort(rows)
	assert.Equal(t, [][]string{{"a", "a"}, {"a", "b"}, {"b", "z"}}, rows)
}

func TestRender_InvalidFormat(t *testing.T) {
	_, err := Render([][]string{{"x"}}, []string{"col"}, Options{Format: Format("invalid")})
	assert.Error(t, err)
}
