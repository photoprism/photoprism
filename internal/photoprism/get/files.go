package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var onceFiles sync.Once

func initFiles() {
	services.Files = photoprism.NewFiles()
}

func Files() *photoprism.Files {
	onceFiles.Do(initFiles)

	return services.Files
}
