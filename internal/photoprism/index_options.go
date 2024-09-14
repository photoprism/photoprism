package photoprism

import "github.com/photoprism/photoprism/internal/entity"

// IndexOptions represents file indexing options.
type IndexOptions struct {
	UID             string
	Action          string
	Path            string
	Rescan          bool
	Convert         bool
	Stack           bool
	FacesOnly       bool
	SkipArchived    bool
	ByteLimit       int64
	ResolutionLimit int
}

// NewIndexOptions returns new index options instance.
func NewIndexOptions(path string, rescan, convert, stack, facesOnly, skipArchived bool) IndexOptions {
	result := IndexOptions{
		UID:             entity.Admin.GetUID(),
		Action:          ActionIndex,
		Path:            path,
		Rescan:          rescan,
		Convert:         convert,
		Stack:           stack,
		FacesOnly:       facesOnly,
		SkipArchived:    skipArchived,
		ByteLimit:       Config().OriginalsByteLimit(),
		ResolutionLimit: Config().ResolutionLimit(),
	}

	return result
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
func IndexOptionsAll() IndexOptions {
	return NewIndexOptions("/", true, true, true, false, true)
}

// IndexOptionsFacesOnly returns new index options for updating faces only.
func IndexOptionsFacesOnly() IndexOptions {
	return NewIndexOptions("/", true, true, true, true, true)
}

// IndexOptionsSingle returns new index options for unstacked, single files.
func IndexOptionsSingle() IndexOptions {
	return NewIndexOptions("/", true, true, false, false, false)
}

// IndexOptionsNone returns new index options with all options set to false.
func IndexOptionsNone() IndexOptions {
	return NewIndexOptions("", false, false, false, false, false)
}
