package photoprism

import (
	"path"

	"github.com/photoprism/photoprism/internal/entity"
)

func FileName(fileRoot, fileName string) string {
	switch fileRoot {
	case entity.RootSidecar:
		return path.Join(Config().SidecarPath(), fileName)
	case entity.RootImport:
		return path.Join(Config().ImportPath(), fileName)
	case entity.RootExamples:
		return path.Join(Config().ExamplesPath(), fileName)
	default:
		return path.Join(Config().OriginalsPath(), fileName)
	}
}
