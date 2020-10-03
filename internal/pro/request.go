package pro

import "runtime"

var ApiURL = "https://api.photoprism.pro/v1/hello"

// photoprism.pro api credentials request incl basic runtime specs for statistical evaluation.
type Request struct {
	ClientVersion string `json:"ClientVersion"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
}

// NewRequest creates a new photoprism.pro key request instance.
func NewRequest(version string) *Request {
	return &Request{
		ClientVersion: version,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
	}
}
