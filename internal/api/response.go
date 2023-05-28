package api

import "net/http"

// Response represents a server status response.
type Response struct {
	Code    int    `json:"code"`
	Err     string `json:"error,omitempty"`
	Msg     string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

// NewResponse creates a new server status response.
func NewResponse(code int, err error, details string) Response {
	if err == nil {
		return Response{Code: http.StatusOK, Msg: "OK", Details: details}
	}
	return Response{Code: code, Err: err.Error(), Details: details}
}
