package customize

type DownloadName string

const (
	DownloadNameFile     DownloadName = "file"
	DownloadNameOriginal DownloadName = "original"
	DownloadNameShare    DownloadName = "share"
)

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
