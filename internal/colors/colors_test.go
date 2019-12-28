package colors

import (
	"testing"
)

func TestColors_List(t *testing.T) {
	allColors := All.List()

	t.Logf("colors: %+v", allColors)
}
