package util

import (
	"net/http"
	"os"
)

func MimeType(filename string) string {
	handle, err := os.Open(filename)

	if err != nil {
		log.Error(err.Error())
		return ""
	}

	defer handle.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = handle.Read(buffer)

	if err != nil {
		log.Errorf("could not read file to determine mime type: %s", filename)
		return ""
	}

	return http.DetectContentType(buffer)
}
