package report

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		cols := []string{"Col1", "Col2"}
		rows := [][]string{
			{"foo", "bar" + strings.Repeat(", abc", 30)},
			{"bar", "b & a | z"}}
		result := Table(rows, cols, false)
		assert.Contains(t, result, "| bar  | b & a | z                      |")
	})
	t.Run("Markdown", func(t *testing.T) {
		cols := []string{"Col1", "Col2"}
		rows := [][]string{
			{"foo", "bar" + strings.Repeat(", abc", 30)},
			{"bar", "b & a | z"}}
		result := Table(rows, cols, true)
		// fmt.Println(result)
		assert.Contains(t, result, "| bar  | b & a \\| z")
	})
}
