package meta

import (
	"time"
)

// Data represents image meta data.
type Data struct {
	UniqueID     string        `meta:"ImageUniqueID"`
	TakenAt      time.Time     `meta:"DateTimeOriginal,CreateDate,MediaCreateDate"`
	TakenAtLocal time.Time     `meta:"DateTimeOriginal,CreateDate,MediaCreateDate"`
	Duration     time.Duration `meta:"Duration,MediaDuration"`
	TimeZone     string        `meta:"-"`
	Title        string        `meta:"Title"`
	Subject      string        `meta:"Subject"`
	Keywords     string        `meta:"Keywords"`
	Comment      string        `meta:"-"`
	Artist       string        `meta:"Artist,Creator"`
	Description  string        `meta:"Description"`
	Copyright    string        `meta:"Rights,Copyright"`
	CameraMake   string        `meta:"CameraMake,Make"`
	CameraModel  string        `meta:"CameraModel,Model"`
	CameraOwner  string        `meta:"OwnerName"`
	CameraSerial string        `meta:"SerialNumber"`
	LensMake     string        `meta:"LensMake"`
	LensModel    string        `meta:"Lens,LensModel"`
	Flash        bool          `meta:"-"`
	FocalLength  int           `meta:"-"`
	Exposure     string        `meta:"ExposureTime"`
	Aperture     float32       `meta:"ApertureValue"`
	FNumber      float32       `meta:"FNumber"`
	Iso          int           `meta:"ISO"`
	GPSPosition  string        `meta:"GPSPosition"`
	Lat          float32       `meta:"-"` // TODO
	Lng          float32       `meta:"-"` // TODO
	Altitude     int           `meta:"-"`
	Width        int           `meta:"ImageWidth"`
	Height       int           `meta:"ImageHeight"`
	Orientation  int           `meta:"-"`
	All          map[string]string
}
