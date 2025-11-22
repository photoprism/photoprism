package customize

// DownloadName represents the naming mode used when exporting or downloading files.
type DownloadName = string

const (
	// DownloadNameFile keeps the original file name when downloading.
	DownloadNameFile DownloadName = "file"
	// DownloadNameOriginal restores the original filename from metadata when available.
	DownloadNameOriginal DownloadName = "original"
	// DownloadNameShare uses the public share identifier for downloaded files.
	DownloadNameShare DownloadName = "share"
)

// DownloadNameDefault is the default naming mode for downloads.
var DownloadNameDefault = DownloadNameFile

// DownloadSettings represents content download settings.
type DownloadSettings struct {
	Name         DownloadName `json:"name" yaml:"Name"`
	Disabled     bool         `json:"disabled" yaml:"Disabled"`
	Originals    bool         `json:"originals" yaml:"Originals"`
	MediaRaw     bool         `json:"mediaRaw" yaml:"MediaRaw"`
	MediaSidecar bool         `json:"mediaSidecar" yaml:"MediaSidecar"`
}

// NewDownloadSettings creates download settings with defaults.
func NewDownloadSettings() DownloadSettings {
	return DownloadSettings{
		Name:         DownloadNameDefault,
		Disabled:     false,
		Originals:    true,
		MediaRaw:     false,
		MediaSidecar: false,
	}
}
