package photoprism

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

// IndexOptions represents file indexing options.
type IndexOptions struct {
	UID             string
	Action          string
	Path            string
	Rescan          bool
	Convert         bool
	Stack           bool
	FacesOnly       bool
	DetectFaces     bool // Detect primary-file faces during indexing.
	DetectNsfw      bool // Flag sensitive content while importing/updating photos.
	GenerateLabels  bool // Generate automatic vision labels for newly indexed files.
	ImportFaceTags  bool // Import face regions and names from XMP metadata.
	SkipArchived    bool
	ByteLimit       int64
	ResolutionLimit int
}

// NewIndexOptions returns new index options instance.
// The config pointer provides the scheduling defaults for detection and labeling features.
func NewIndexOptions(path string, rescan, convert, stack, facesOnly, skipArchived bool, c *config.Config) IndexOptions {
	result := IndexOptions{
		UID:          entity.Admin.GetUID(),
		Action:       ActionIndex,
		Path:         path,
		Rescan:       rescan,
		Convert:      convert,
		Stack:        stack,
		FacesOnly:    facesOnly,
		DetectFaces:  facesOnly,
		SkipArchived: skipArchived,
	}

	if c != nil {
		var facesRunType vision.RunType

		if rescan || facesOnly {
			// Manual runs bypass CPU-based scheduling so face detection still runs during forced rescans.
			facesRunType = vision.RunManual
		} else {
			facesRunType = vision.RunOnIndex
		}

		result.DetectFaces = c.VisionModelShouldRun(vision.ModelTypeFace, facesRunType)
		result.DetectNsfw = !facesOnly && c.VisionModelShouldRun(vision.ModelTypeNsfw, vision.RunOnIndex)
		result.GenerateLabels = !facesOnly && c.VisionModelShouldRun(vision.ModelTypeLabels, vision.RunOnIndex)

		result.ImportFaceTags = c.XMPFaces()

		result.ByteLimit = c.OriginalsLimitBytes()
		result.ResolutionLimit = c.ResolutionLimit()
	}

	return result
}

// ResolveIndexPath resolves the index subpath under originalsPath and returns the
// absolute directory to scan. Empty, ".", and "/" resolve to the originals root;
// paths that would escape the originals directory via parent-directory traversal
// or absolute/volume references are rejected.
func ResolveIndexPath(originalsPath, subPath string) (string, error) {
	sub := strings.Trim(filepath.ToSlash(subPath), "/")

	if sub == "" || sub == "." {
		return filepath.Clean(originalsPath), nil
	}

	return fs.SafeJoin(originalsPath, sub)
}

// SkipUnchanged checks if unchanged media files should be skipped.
func (o *IndexOptions) SkipUnchanged() bool {
	return !o.Rescan
}

// SetUser sets the user who performs the index operation.
func (o *IndexOptions) SetUser(user *entity.User) *IndexOptions {
	if o != nil && user != nil {
		o.UID = user.GetUID()
	}

	return o
}

// IndexOptionsAll returns new index options with all options set to true.
func IndexOptionsAll(c *config.Config) IndexOptions {
	return NewIndexOptions("/", true, true, true, false, true, c)
}

// IndexOptionsFacesOnly returns new index options for updating faces only.
func IndexOptionsFacesOnly(c *config.Config) IndexOptions {
	return NewIndexOptions("/", true, true, true, true, true, c)
}

// IndexOptionsSingle returns new index options for unstacked, single files.
func IndexOptionsSingle(c *config.Config) IndexOptions {
	return NewIndexOptions("/", true, true, false, false, false, c)
}

// IndexOptionsNone returns new index options with all options set to false.
func IndexOptionsNone(c *config.Config) IndexOptions {
	return NewIndexOptions("", false, false, false, false, false, c)
}
