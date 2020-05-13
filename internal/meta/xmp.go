package meta

import (
	"fmt"
)

// XMP parses an XMP file and returns a Data struct.
func XMP(fileName string) (data Data, err error) {
	err = data.XMP(fileName)

	return data, err
}

// XMP parses an XMP file and returns a Data struct.
func (data *Data) XMP(fileName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("meta: %s", e)
		}
	}()

	doc := XmpDocument{}

	if err := doc.Load(fileName); err != nil {
		return err
	}

	if doc.Title() != "" {
		data.Title = doc.Title()
	}

	if doc.Artist() != "" {
		data.Artist = doc.Artist()
	}

	if doc.Description() != "" {
		data.Description = doc.Description()
	}

	if doc.Copyright() != "" {
		data.Copyright = doc.Copyright()
	}

	if doc.CameraMake() != "" {
		data.CameraMake = doc.CameraMake()
	}

	if doc.CameraModel() != "" {
		data.CameraModel = doc.CameraModel()
	}

	if doc.LensModel() != "" {
		data.LensModel = doc.LensModel()
	}

	return nil
}
