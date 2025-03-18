package thumb

import (
	"reflect"
)

// Viewer represents thumbnail URLs for the photo/video viewer.
type Viewer struct {
	Fit720  *Thumb `json:"fit_720"`
	Fit1280 *Thumb `json:"fit_1280"`
	Fit1920 *Thumb `json:"fit_1920"`
	Fit2560 *Thumb `json:"fit_2560"`
	Fit4096 *Thumb `json:"fit_4096"`
	Fit5120 *Thumb `json:"fit_5120"`
	Fit7680 *Thumb `json:"fit_7680"`
}

// ViewerThumbs creates and returns a Viewer struct pointer with the required thumbnail URLs for the photo/video viewer.
func ViewerThumbs(fileWidth, fileHeight int, fileHash, contentUri, previewToken string) *Viewer {
	thumbs := &Viewer{}

	// Get Viewer struct fields.
	fields := reflect.ValueOf(thumbs).Elem()

	// Remember the maximum allowed size and the maximum actual size.
	var maxSize = MaxSize()
	var largestSize Size

	// Iterate through all Viewer struct fields and set the best matching thumb size.
	for i := 0; i < fields.NumField(); i++ {
		thumb := fields.Field(i)

		// For simplicity, JSON value name is the same as the thumbnail size name.
		size := Name(fields.Type().Field(i).Tag.Get("json"))

		if size == "" {
			continue
		}

		s := Sizes[size]

		// Make sure not to process an invalid size.
		if s.Name == "" {
			continue
		}

		// Remember this as the largest size needed if the original size is smaller than the thumb size.
		if largestSize.Name == "" && (s.Width >= fileWidth && s.Height >= fileHeight || s.Width >= maxSize) {
			largestSize = s
		}

		// Set the field value to the current size or the maximum size, if any.
		if largestSize.Name != "" {
			thumb.Set(reflect.ValueOf(New(fileWidth, fileHeight, fileHash, largestSize, contentUri, previewToken)))
		} else {
			thumb.Set(reflect.ValueOf(New(fileWidth, fileHeight, fileHash, s, contentUri, previewToken)))
		}
	}

	return thumbs
}
