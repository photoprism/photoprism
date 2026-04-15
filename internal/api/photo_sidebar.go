package api

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

// PhotoSidebar is the reduced JSON shape returned by GetPhoto to sessions
// whose role does not satisfy acl.SidebarFullAccess. It drops Camera, ISO,
// Exposure, Lens, F-Number, Focal Length, Place, Altitude, Labels, Albums,
// People, Subject, Notes, Keywords, Path, Artist, Copyright, and License so
// those values cannot leak through direct API calls regardless of any
// client-side flags.
type PhotoSidebar struct {
	UID          string        `json:"UID"`
	Type         string        `json:"Type"`
	Title        string        `json:"Title"`
	Caption      string        `json:"Caption"`
	TakenAt      time.Time     `json:"TakenAt"`
	TakenAtLocal time.Time     `json:"TakenAtLocal"`
	TimeZone     string        `json:"TimeZone"`
	Year         int           `json:"Year"`
	Month        int           `json:"Month"`
	Day          int           `json:"Day"`
	Lat          float64       `json:"Lat"`
	Lng          float64       `json:"Lng"`
	Duration     time.Duration `json:"Duration,omitempty"`
	// TODO(#1307): wrap in a FileSidebar DTO that drops FileName, OriginalName,
	// and FileRoot — the current shape still exposes the stored path through
	// Files[*].Name. Left as-is for now because the lightbox, download, and
	// thumbnail helpers all read from entity.File directly.
	Files []entity.File `json:"Files"`
}

// BuildPhotoSidebar returns a PhotoSidebar DTO populated from the given
// photo entity. Preloaded relations (Details, Camera, Lens, Cell, Place,
// Labels, Albums) are intentionally dropped.
func BuildPhotoSidebar(p entity.Photo) PhotoSidebar {
	return PhotoSidebar{
		UID:          p.PhotoUID,
		Type:         p.PhotoType,
		Title:        p.PhotoTitle,
		Caption:      p.PhotoCaption,
		TakenAt:      p.TakenAt,
		TakenAtLocal: p.TakenAtLocal,
		TimeZone:     p.TimeZone,
		Year:         p.PhotoYear,
		Month:        p.PhotoMonth,
		Day:          p.PhotoDay,
		Lat:          p.PhotoLat,
		Lng:          p.PhotoLng,
		Duration:     p.PhotoDuration,
		Files:        p.Files,
	}
}
