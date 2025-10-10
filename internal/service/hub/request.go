package hub

import (
	"runtime"
)

// Request represents basic environment specs for debugging.
type Request struct {
	ClientVersion string `json:"ClientVersion"`
	ClientSerial  string `json:"ClientSerial"`
	ClientOS      string `json:"ClientOS"`
	ClientArch    string `json:"ClientArch"`
	ClientCPU     int    `json:"ClientCPU"`
	ClientEnv     string `json:"ClientEnv"`
	ClientOpt     string `json:"ClientOpt"`
	PartnerID     string `json:"PartnerID"`
	ApiToken      string `json:"ApiToken"`
}

// ClientOpt hooks let tests and extensions append optional context information
// to Hub requests; callers may replace the function to emit custom strings.
var ClientOpt = func() string {
	return ""
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
		ClientOpt:     ClientOpt(),
		PartnerID:     partnerId,
		ApiToken:      token,
	}
}
