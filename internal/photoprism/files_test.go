package photoprism

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFiles_Ignore(t *testing.T) {
	files := NewFiles()

	if err := files.Init(); err != nil {
		t.Fatal(err)
	}

	assert.True(t, files.Ignore("exampleFileName.jpg", time.Unix(1583460411, 0), false))
	assert.False(t, files.Ignore("exampleFileName.jpg", time.Unix(1583460412, 0), false))
	assert.True(t, files.Ignore("exampleFileName.jpg", time.Unix(1583460412, 0), false))
	assert.False(t, files.Ignore("exampleFileName.jpg", time.Unix(1583460412, 0), true))
	assert.False(t, files.Ignore("exampleFileName.jpg", time.Unix(500, 0), false))
	assert.True(t, files.Ignore("exampleFileName.jpg", time.Unix(500, 0), false))
	assert.False(t, files.Ignore("new-file.jpg", time.Unix(1583460000, 1), false))
	assert.True(t, files.Ignore("new-file.jpg", time.Unix(1583460000, 2), false))
	assert.False(t, files.Ignore("new-file.jpg", time.Unix(1583460001, 2), false))
	assert.False(t, files.Ignore("new-file.jpg", time.Unix(1583460001, 2), true))
	assert.True(t, files.Ignore("new-file.jpg", time.Unix(1583460001, 2), false))
	assert.False(t, files.Ignore("new-file.jpg", time.Unix(501, 0), false))
}
