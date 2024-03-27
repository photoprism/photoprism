package media

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/gabriel-vasile/mimetype"
)

func Base64(buf *bytes.Buffer) string {
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	if encoded == "" {
		return ""
	}

	var mimeType string

	mime, err := mimetype.DetectReader(buf)

	if err != nil {
		mimeType = "application/octet-stream"
	} else {
		mimeType = mime.String()
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
}
