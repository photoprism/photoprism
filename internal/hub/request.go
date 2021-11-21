package hub

import (
	"net/url"
	"runtime"
)

var ServiceURL = "https://hub.photoprism.app/v1/hello"

// Request represents basic environment specs for debugging.
type Request struct {
	ClientVersion string `json:"ClientVersion"`
	ClientSerial  string `json:"ClientSerial"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
}

// NewRequest creates a new backend key request instance.
func NewRequest(version, serial string) *Request {
	return &Request{
		ClientVersion: version,
		ClientSerial:  serial,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
	}
}

// ApiHost returns the backend host name.
func ApiHost() string {
	u, err := url.Parse(ServiceURL)

	if err != nil {
		log.Warn(err)
		return ""
	}

	return u.Host
}
