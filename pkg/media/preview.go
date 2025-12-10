package media

import (
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/fs"
)

// PreviewFileTypes lists MIME types eligible for preview generation.
var PreviewFileTypes = []string{fs.ImageJpeg.String(), fs.ImagePng.String()}

// PreviewExpr is a SQL expression containing allowed preview MIME types.
var PreviewExpr = gorm.Expr("'" + strings.Join(PreviewFileTypes, "','") + "'")
