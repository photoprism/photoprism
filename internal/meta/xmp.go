package meta

import (
	"fmt"
)

// XMP parses an XMP file and returns a Data struct.
func XMP(filename string) (data Data, err error) {
	defer func() {
		if e := recover(); e != nil {
			data = Data{}
			err = fmt.Errorf("meta: %s", e)
		}
	}()

	doc := XmpDocument{}

	if err := doc.Load(filename); err != nil {
		return data, err
	}

	data.Title = doc.Title()
	data.Artist = doc.Artist()
	data.Description = doc.Description()
	data.Copyright = doc.Copyright()
	data.CameraMake = doc.CameraMake()
	data.CameraModel = doc.CameraModel()
	data.LensModel = doc.LensModel()

	return data, nil
}
