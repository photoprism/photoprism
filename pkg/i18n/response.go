package i18n

import "strings"

// Response represents an i18n-aware response payload.
//
// Error/Message carry the server-rendered string in the instance locale (a fallback for non-browser
// consumers); MessageID and MessageParams carry the untranslated source string and its parameters
// so the Web UI can render the message in each user's current UI locale.
type Response struct {
	Code          int    `json:"code"`
	Error         string `json:"error,omitempty"`
	Message       string `json:"message,omitempty"`
	MessageID     string `json:"messageId,omitempty"`
	MessageParams []any  `json:"messageParams,omitempty"`
	Details       string `json:"details,omitempty"`
}

// String returns the response message as string.
func (r Response) String() string {
	if r.Error != "" {
		return r.Error
	}

	return r.Message
}

// LowerString returns the lowercased message string.
func (r Response) LowerString() string {
	return strings.ToLower(r.String())
}

// ErrorString returns the error message as string.
func (r Response) ErrorString() string {
	return r.Error
}

// Success reports whether the response code indicates success (2xx).
func (r Response) Success() bool {
	return r.Error == "" && r.Code < 400
}

// NewResponse builds a Response with the given code, message ID, and optional parameters.
// It carries the untranslated source string (MessageID) and MessageParams so the frontend can render
// the message in the current UI locale, in addition to the server-rendered Error/Message fallback.
func NewResponse(code int, id Message, params ...any) Response {
	r := Response{Code: code, MessageID: Source(id), MessageParams: params}

	if code < 400 {
		r.Message = Msg(id, params...)
	} else {
		r.Error = Msg(id, params...)
	}

	return r
}
