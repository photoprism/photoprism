package media

import (
	"testing"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestPreviewFileTypes(t *testing.T) {
	assert.Equal(t, []string{"jpg", "png"}, PreviewFileTypes)
}

func TestIsPreview(t *testing.T) {
	assert.Equal(t, gorm.Expr("'jpg','png'"), PreviewExpr)
}
