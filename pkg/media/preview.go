package media

import (
	"strings"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/fs"
)

var PreviewFileTypes = []string{fs.ImageJPEG.String(), fs.ImagePNG.String()}
var PreviewExpr = gorm.Expr("'" + strings.Join(PreviewFileTypes, "','") + "'")
