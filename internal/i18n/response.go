package i18n

import "strings"

type Response struct {
	Code int    `json:"code"`
	Err  string `json:"error,omitempty"`
	Msg  string `json:"success,omitempty"`
}

func (r Response) String() string {
	if r.Err != "" {
		return r.Err
	} else {
		return r.Msg
	}
}

func (r Response) LowerString() string {
	return strings.ToLower(r.String())
}

func (r Response) Error() string {
	return r.Err
}

func NewResponse(code int, id Message, params ...interface{}) Response {
	if code < 400 {
		return Response{Code: code, Msg: Msg(id, params...)}
	} else {
		return Response{Code: code, Err: Msg(id, params...)}
	}
}
