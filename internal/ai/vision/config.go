package vision

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

var (
	CachePath             = ""
	ModelsPath            = ""
	DownloadUrl           = ""
	ServiceUri            = ""
	ServiceKey            = ""
	ServiceTimeout        = 10 * time.Minute
	ServiceMethod         = http.MethodPost
	ServiceFileScheme     = scheme.Data
	ServiceRequestFormat  = ApiFormatVision
	ServiceResponseFormat = ApiFormatVision
	DefaultResolution     = 224
)

// SetCachePath updates the cache path.
func SetCachePath(dir string) {
	if dir = fs.Abs(dir); dir == "" {
		return
	}

	CachePath = dir
}

// GetCachePath returns the cache path.
func GetCachePath() string {
	if CachePath != "" {
		return CachePath
	}

	CachePath = fs.Abs("../../../storage/cache")

	return CachePath
}

// SetModelsPath updates the model assets path.
func SetModelsPath(dir string) {
	if dir = fs.Abs(dir); dir == "" {
		return
	}

	ModelsPath = dir
}

// GetModelsPath returns the model assets path, or an empty string if not configured or found.
func GetModelsPath() string {
	if ModelsPath != "" {
		return ModelsPath
	}

	assetsPath := fs.Abs("../../../assets")

	if dir := filepath.Join(assetsPath, "models"); fs.PathExists(dir) {
		ModelsPath = dir
	} else if fs.PathExists(assetsPath) {
		ModelsPath = assetsPath
	}

	return ModelsPath
}

func GetModelPath(name string) string {
	return filepath.Join(GetModelsPath(), clean.Path(clean.TypeLowerUnderscore(name)))
}

func GetNasnetModelPath() string {
	return GetModelPath(NasnetModel.Name)
}

func GetFacenetModelPath() string {
	return GetModelPath(FacenetModel.Name)
}

func GetNsfwModelPath() string {
	return GetModelPath(NsfwModel.Name)
}
