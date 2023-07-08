package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	form := &SearchPhotos{}
	rows, cols := Report(form)
	assert.Contains(t, rows[5][0], "name")
	assert.Contains(t, rows[5][1], "string")
	assert.Contains(t, rows[5][2], "name:\"IMG_9831-112*\"")

	assert.Contains(t, cols, "Examples")
}
