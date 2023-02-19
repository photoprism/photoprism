package form

import (
	"encoding/json"
	"io"
	"strings"
)

// AsJson returns the form data as a JSON string or an empty string in case of error.
func AsJson(frm interface{}) string {
	s, _ := json.Marshal(frm)

	return string(s)
}

// AsReader returns the form data as io.Reader, e.g. for use in tests.
func AsReader(frm interface{}) io.Reader {
	return strings.NewReader(AsJson(frm))
}
