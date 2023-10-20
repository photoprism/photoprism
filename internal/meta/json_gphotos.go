package meta

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
	"gopkg.in/photoprism/go-tz.v2/tz"
)

type GPhoto struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Views       int    `json:"imageViews,string"`
	Geo         GGeo   `json:"geoData"`
	TakenAt     GTime  `json:"photoTakenTime"`
	CreatedAt   GTime  `json:"creationTime"`
	UpdatedAt   GTime  `json:"modificationTime"`
}

func (m GPhoto) SanitizedTitle() string {
	return SanitizeTitle(m.Title)
}

func (m GPhoto) SanitizedDescription() string {
	return SanitizeDescription(m.Description)
}

type GMeta struct {
	Album GAlbum `json:"albumData"`
}

type GAlbum struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Access      string `json:"access"`
	Location    string `json:"location"`
	Date        GTime  `json:"date"`
	Geo         GGeo   `json:"geoData"`
}

func (m GAlbum) Exists() bool {
	return m.Title != ""
}

type GGeo struct {
	Lat      float64 `json:"latitude"`
	Lng      float64 `json:"longitude"`
	Altitude float64 `json:"altitude"`
}

func (m GGeo) Exists() bool {
	return m.Lat != 0.0 && m.Lng != 0.0
}

type GTime struct {
	Unix      int64  `json:"timestamp,string"`
	Formatted string `json:"formatted"`
}

func (m GTime) Exists() bool {
	return m.Unix > 0
}

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

	if s := p.SanitizedTitle(); s != "" && data.Title == "" {
		data.Title = s
	}

	if s := p.SanitizedDescription(); s != "" && data.Description == "" {
		data.Description = s
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
			data.Lat = float32(p.Geo.Lat)
			data.Lng = float32(p.Geo.Lng)
		}

		if data.Altitude == 0 {
			data.Altitude = p.Geo.Altitude
		}
	}

	// Set time zone and calculate UTC time.
	if data.Lat != 0 && data.Lng != 0 {
		zones, err := tz.GetZone(tz.Point{
			Lat: float64(data.Lat),
			Lon: float64(data.Lng),
		})

		if err == nil && len(zones) > 0 {
			data.TimeZone = zones[0]
		}

		if !data.TakenAtLocal.IsZero() {
			if loc := txt.TimeZone(data.TimeZone); loc == nil {
				log.Warnf("metadata: invalid time zone %s (gphotos)", data.TimeZone)
			} else if tl, err := time.ParseInLocation("2006:01:02 15:04:05", data.TakenAtLocal.Format("2006:01:02 15:04:05"), loc); err == nil {
				data.TakenAt = tl.UTC().Truncate(time.Second)
			} else {
				log.Errorf("metadata: %s (gphotos)", err.Error()) // this should never happen
			}
		}
	}

	return nil
}
