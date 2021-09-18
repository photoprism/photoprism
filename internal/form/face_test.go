package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFace(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var m = struct {
			FaceHidden bool   `json:"Hidden"`
			SubjUID    string `json:"SubjUID"`
		}{
			FaceHidden: true,
			SubjUID:    "jqzmd5q3b8o2yxu7",
		}

		f, err := NewFace(m)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, f.FaceHidden)
		assert.Equal(t, "jqzmd5q3b8o2yxu7", f.SubjUID)
	})
}
