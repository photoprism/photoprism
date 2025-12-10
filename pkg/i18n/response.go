package i18n

import "strings"

// Response represents an i18n-aware response payload.
type Response struct {
	Code    int    `json:"code"`
	Err     string `json:"error,omitempty"`
	Msg     string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
}

func (r Response) String() string {
	if r.Err != "" {
		return r.Err
	} else {
		return r.Msg
	}
}

// LowerString returns the lowercased message string.
func (r Response) LowerString() string {
	return strings.ToLower(r.String())
}

func (r Response) Error() string {
	return r.Err
}

// Success reports whether the response code indicates success (2xx).
func (r Response) Success() bool {
	return r.Err == "" && r.Code < 400
}

// NewResponse builds a Response with the given code, message ID, and optional parameters.
func NewResponse(code int, id Message, params ...interface{}) Response {
	if code < 400 {
		return Response{Code: code, Msg: Msg(id, params...)}
	} else {
		return Response{Code: code, Err: Msg(id, params...)}
	}
}
