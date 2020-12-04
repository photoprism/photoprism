package hub

import (
	"net/url"
	"runtime"
)

var ServiceURL = "https://hub.photoprism.app/v1/hello"

// Backend api credentials request incl basic runtime specs.
type Request struct {
	ClientVersion string `json:"ClientVersion"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
}

// NewRequest creates a new hub key request instance.
func NewRequest(version string) *Request {
	return &Request{
		ClientVersion: version,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
	}
}

// ApiHost returns the full API URL host name.
func ApiHost() string {
	u, err := url.Parse(ServiceURL)

	if err != nil {
		log.Warn(err)
		return ""
	}

	return u.Host
}
