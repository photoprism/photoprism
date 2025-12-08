package meta

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/pkg/time/tz"
)

// GPhoto represents the photo-level fields exported by Google Photos JSON sidecars.
type GPhoto struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Views       int    `json:"imageViews,string"`
	Geo         GGeo   `json:"geoData"`
	TakenAt     GTime  `json:"photoTakenTime"`
	CreatedAt   GTime  `json:"creationTime"`
	UpdatedAt   GTime  `json:"modificationTime"`
}

// GetTitle returns the sanitized Google Photos title.
func (m GPhoto) GetTitle() string {
	return SanitizeTitle(m.Title)
}

// GetCaption returns the sanitized Google Photos description.
func (m GPhoto) GetCaption() string {
	return SanitizeCaption(m.Description)
}

// GMeta wraps album metadata embedded in Google Photos sidecars.
type GMeta struct {
	Album GAlbum `json:"albumData"`
}

// GAlbum contains album-level information from Google Photos exports.
type GAlbum struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Access      string `json:"access"`
	Location    string `json:"location"`
	Date        GTime  `json:"date"`
	Geo         GGeo   `json:"geoData"`
}

// Exists reports whether the album entry has meaningful data.
func (m GAlbum) Exists() bool {
	return m.Title != ""
}

// GGeo holds geolocation data provided by Google Photos.
type GGeo struct {
	Lat      float64 `json:"latitude"`
	Lng      float64 `json:"longitude"`
	Altitude float64 `json:"altitude"`
}

// Exists reports whether the geolocation entry has usable coordinates.
func (m GGeo) Exists() bool {
	return m.Lat != 0.0 && m.Lng != 0.0
}

// GTime stores Unix timestamps used in Google Photos metadata.
type GTime struct {
	Unix      int64  `json:"timestamp,string"`
	Formatted string `json:"formatted"`
}

// Exists reports whether the timestamp is set.
func (m GTime) Exists() bool {
	return m.Unix > 0
}

// Time returns the timestamp as a UTC time.Time.
func (m GTime) Time() time.Time {
	return time.Unix(m.Unix, 0).UTC()
}

// GMeta parses JSON sidecar data as created by Google Photos.
func (data *Data) GMeta(jsonData []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s (gmeta panic)\nstack: %s", e, debug.Stack())
		}
	}()

	p := GMeta{}

	if err := json.Unmarshal(jsonData, &p); err != nil {
		return err
	}

	if p.Album.Exists() {
		data.Albums = append(data.Albums, p.Album.Title)
	}

	return nil
}

// GPhoto parses JSON photo sidecar data as created by Google Photos.
func (data *Data) GPhoto(jsonData []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s (gphoto panic)\nstack: %s", e, debug.Stack())
		}
	}()

	p := GPhoto{}

	if err := json.Unmarshal(jsonData, &p); err != nil {
		return err
	}

	if s := p.GetTitle(); s != "" && data.Title == "" {
		data.Title = s
	}

	if s := p.GetCaption(); s != "" && data.Caption == "" {
		data.Caption = s
	}

	if p.Views > 0 && data.Views == 0 {
		data.Views = p.Views
	}

	if p.TakenAt.Exists() {
		if data.TakenAt.IsZero() {
			data.TakenAt = p.TakenAt.Time()
		}

		if data.TakenAtLocal.IsZero() {
			data.TakenAtLocal = p.TakenAt.Time()
		}
	}

	if p.Geo.Exists() {
		if data.Lat == 0 && data.Lng == 0 {
			data.Lat = p.Geo.Lat
			data.Lng = p.Geo.Lng
		}

		if data.Altitude == 0 {
			data.Altitude = p.Geo.Altitude
		}
	}

	// Set time zone and calculate UTC time.
	if data.Lat != 0 && data.Lng != 0 {
		if zone := tz.Position(data.Lat, data.Lng); zone != "" {
			data.TimeZone = zone
		}

		if loc := tz.Find(data.TimeZone); !data.TakenAtLocal.IsZero() {
			if tl, locErr := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), loc); locErr == nil {
				data.TakenAt = tl.UTC().Truncate(time.Second)
			} else {
				log.Errorf("metadata: %s (gphotos)", locErr.Error()) // this should never happen
			}
		}
	}

	// Normalize time zone name.
	data.TimeZone = tz.Name(data.TimeZone)

	return nil
}
