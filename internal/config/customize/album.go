package customize

import (
	"github.com/photoprism/photoprism/internal/entity/sortby"
)

// AlbumsSettings represents album defaults and preferences.
type AlbumsSettings struct {
	Order    AlbumsOrder      `json:"order" yaml:"Order"`
	Download DownloadSettings `json:"download" yaml:"Download"`
}

// AlbumsOrder represents default album sort orders.
type AlbumsOrder struct {
	Album  string `json:"album" yaml:"Album,omitempty"`
	Folder string `json:"folder" yaml:"Folder,omitempty"`
	Moment string `json:"moment" yaml:"Moment,omitempty"`
	State  string `json:"state" yaml:"State,omitempty"`
	Month  string `json:"month" yaml:"Month,omitempty"`
}

// NewAlbumSettings creates album settings with defaults.
func NewAlbumSettings() AlbumsSettings {
	return AlbumsSettings{
		Order: AlbumsOrder{
			Album:  sortby.Oldest,
			Folder: sortby.Added,
			Moment: sortby.Oldest,
			State:  sortby.Newest,
			Month:  sortby.Oldest,
		},
		Download: DownloadSettings{
			Name:         DownloadNameShare,
			Disabled:     false,
			Originals:    true,
			MediaRaw:     false,
			MediaSidecar: false,
		},
	}
}
