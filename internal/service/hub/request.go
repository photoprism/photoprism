package hub

import (
	"net/url"
	"runtime"
)

var ServiceURL = "https://my.photoprism.app/v1/hello"

// Request represents basic environment specs for debugging.
type Request struct {
	ClientVersion string `json:"ClientVersion"`
	ClientSerial  string `json:"ClientSerial"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
	ClientEnv     string `json:"ClientEnv"`
	PartnerID     string `json:"PartnerID"`
	ApiToken      string `json:"ApiToken"`
}

// NewRequest creates a new backend key request instance.
func NewRequest(version, serial, env, partnerId, token string) *Request {
	return &Request{
		ClientVersion: version,
		ClientSerial:  serial,
		ClientOS:      runtime.GOOS,
		ClientArch:    runtime.GOARCH,
		ClientCPU:     runtime.NumCPU(),
		ClientEnv:     env,
		PartnerID:     partnerId,
		ApiToken:      token,
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
