package meta

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/projection"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Common MIME types used to detect video contexts in ExifTool sidecars.
const (
	MimeVideoMp4  = "video/mp4"
	MimeQuicktime = "video/quicktime"
)

// jsonAt returns the array element at index i, or an empty result when out of
// range. ExifTool emits a single region as scalars and multiple regions as
// index-parallel arrays; gjson's Array() returns a one-element slice for a
// scalar, so callers can zip both shapes uniformly.
func jsonAt(a []gjson.Result, i int) gjson.Result {
	if i >= 0 && i < len(a) {
		return a[i]
	}
	return gjson.Result{}
}

// jsonFaceName returns the first name available for an ExifTool face region.
func jsonFaceName(j gjson.Result, names, titles, extensions, references []gjson.Result, i int) string {
	if name := jsonAt(names, i).String(); name != "" {
		return name
	} else if title := jsonAt(titles, i).String(); title != "" {
		return title
	} else if extension := jsonAt(extensions, i).String(); extension != "" {
		return extension
	}

	if jsonAt(references, i).String() != "Iptc4xmpExt:PersonInImage" {
		return ""
	}

	names = j.Get("PersonInImage").Array()
	unique := make([]string, 0, len(names))
	for _, result := range names {
		if name := strings.TrimSpace(result.String()); name != "" {
			unique = append(unique, name)
		}
	}
	if unique = txt.UniqueNames(unique); len(unique) == 1 {
		return unique[0]
	}

	return ""
}

// parseExiftoolFaces extracts supported XMP face regions from ExifTool JSON.
func parseExiftoolFaces(j gjson.Result, options FaceOptions) []Face {
	var faces []Face

	// MWG-RS regions (XMP-mwg-rs): parallel arrays keyed by index.
	names := j.Get("RegionName").Array()
	titles := j.Get("RegionTitle").Array()
	extensions := j.Get("RegionPersonInImage").Array()
	references := j.Get("RegionSeeAlso").Array()
	types := j.Get("RegionType").Array()
	xs := j.Get("RegionAreaX").Array()
	ys := j.Get("RegionAreaY").Array()
	ws := j.Get("RegionAreaW").Array()
	hs := j.Get("RegionAreaH").Array()
	ds := j.Get("RegionAreaD").Array()
	units := j.Get("RegionAreaUnit").Array()
	rotations := j.Get("RegionRotation").Array()

	appliedW := int(j.Get("RegionAppliedToDimensionsW").Int())
	appliedH := int(j.Get("RegionAppliedToDimensionsH").Int())
	if appliedW <= 0 || appliedH <= 0 {
		appliedW, appliedH = options.Width, options.Height
	}

	// The coordinate arrays define how many regions exist; a missing name is
	// allowed (unnamed region), so it does not bound the count.
	n := len(xs)
	for _, l := range []int{len(ys)} {
		if l < n {
			n = l
		}
	}

	for i := 0; i < n; i++ {
		if !strings.EqualFold(jsonAt(types, i).String(), "Face") {
			continue
		}

		wResult, hResult := jsonAt(ws, i), jsonAt(hs, i)
		dResult := jsonAt(ds, i)
		if !dResult.Exists() && (!wResult.Exists() || !hResult.Exists()) {
			continue
		}

		if f, ok := normalizeRegionMWG(
			jsonFaceName(j, names, titles, extensions, references, i),
			float32(jsonAt(xs, i).Float()), float32(jsonAt(ys, i).Float()),
			float32(wResult.Float()), float32(hResult.Float()), float32(dResult.Float()),
			jsonAt(units, i).String(), float32(jsonAt(rotations, i).Float()),
			appliedW, appliedH, options.Orientation,
		); ok {
			faces = append(faces, f)
		}
	}

	// Microsoft MP:RegionInfo (XMP-MP): rectangle string + display name.
	mpNames := j.Get("RegionPersonDisplayName").Array()
	mpRects := j.Get("RegionRectangle").Array()

	for i := 0; i < len(mpRects); i++ {
		if f, ok := normalizeRegionMP(jsonAt(mpNames, i).String(), jsonAt(mpRects, i).String(), options.Orientation); ok {
			faces = append(faces, f)
		}
	}

	// ACDSee regions: DLYArea is the user-adjusted rectangle and takes
	// precedence over the automatic ALGArea.
	acdNames := j.Get("ACDSeeRegionName").Array()
	acdTypes := j.Get("ACDSeeRegionType").Array()
	dlyXs := j.Get("ACDSeeRegionDLYAreaX").Array()
	dlyYs := j.Get("ACDSeeRegionDLYAreaY").Array()
	dlyWs := j.Get("ACDSeeRegionDLYAreaW").Array()
	dlyHs := j.Get("ACDSeeRegionDLYAreaH").Array()
	algXs := j.Get("ACDSeeRegionALGAreaX").Array()
	algYs := j.Get("ACDSeeRegionALGAreaY").Array()
	algWs := j.Get("ACDSeeRegionALGAreaW").Array()
	algHs := j.Get("ACDSeeRegionALGAreaH").Array()

	acdW := int(j.Get("ACDSeeRegionAppliedToDimensionsW").Int())
	acdH := int(j.Get("ACDSeeRegionAppliedToDimensionsH").Int())
	if acdW <= 0 || acdH <= 0 {
		acdW, acdH = options.Width, options.Height
	}

	acdCount := len(acdTypes)
	for i := 0; i < acdCount; i++ {
		if !strings.EqualFold(jsonAt(acdTypes, i).String(), "Face") {
			continue
		}

		xs, ys, ws, hs := dlyXs, dlyYs, dlyWs, dlyHs
		if !jsonAt(xs, i).Exists() {
			xs, ys, ws, hs = algXs, algYs, algWs, algHs
		}
		if !jsonAt(xs, i).Exists() || !jsonAt(ys, i).Exists() || !jsonAt(ws, i).Exists() || !jsonAt(hs, i).Exists() {
			continue
		}

		if f, ok := normalizeRegionMWG(
			jsonAt(acdNames, i).String(),
			float32(jsonAt(xs, i).Float()), float32(jsonAt(ys, i).Float()),
			float32(jsonAt(ws, i).Float()), float32(jsonAt(hs, i).Float()), 0,
			"normalized", 0, acdW, acdH, options.Orientation,
		); ok {
			faces = append(faces, f)
		}
	}

	return DedupFaces(faces)
}

// Exiftool parses JSON sidecar data as created by Exiftool.
func (data *Data) Exiftool(jsonData []byte, originalName string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("metadata: %s (exiftool panic)\nstack: %s", e, debug.Stack())
		}
	}()

	j := gjson.GetBytes(jsonData, "@flatten|@join")

	logName := "json file"

	if originalName != "" {
		logName = clean.Log(filepath.Base(originalName))
	}

	if !j.IsObject() {
		return fmt.Errorf("metadata: data is not an object in %s (exiftool)", logName)
	}

	data.json = make(map[string]string)
	jsonValues := j.Map()

	for key, val := range jsonValues {
		data.json[key] = val.String()
	}

	if fileName, ok := data.json["FileName"]; ok && fileName != "" && originalName != "" && fileName != originalName {
		return fmt.Errorf("metadata: original name %s does not match %s (exiftool)", clean.Log(originalName), clean.Log(fileName))
	} else if fileName != "" && originalName == "" {
		logName = clean.Log(filepath.Base(fileName))
	}

	v := reflect.ValueOf(data).Elem()

	// Iterate through all config fields
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagData := v.Type().Field(i).Tag.Get("meta")

		// Automatically assign values to fields with "flag" tag
		if tagData != "" {
			tagValues := strings.Split(tagData, ",")

			var jsonValue gjson.Result
			var tagValue string

			for _, tagValue = range tagValues {
				if r, ok := jsonValues[tagValue]; !ok {
					continue
				} else if txt.Empty(r.String()) {
					continue
				} else {
					jsonValue = r
					break
				}
			}

			// Skip empty values.
			if !jsonValue.Exists() {
				continue
			}

			switch t := fieldValue.Interface().(type) {
			case time.Time:
				if !fieldValue.IsZero() {
					continue
				}

				if dateTime := txt.ParseTime(jsonValue.String(), ""); !dateTime.IsZero() {
					fieldValue.Set(reflect.ValueOf(dateTime))
				}
			case time.Duration:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.Set(reflect.ValueOf(Duration(jsonValue.String())))
			case int, int64:
				if !fieldValue.IsZero() {
					continue
				}

				if intVal := jsonValue.Int(); intVal != 0 {
					fieldValue.SetInt(intVal)
				} else if intVal = txt.Int64(jsonValue.String()); intVal != 0 {
					fieldValue.SetInt(intVal)
				}
			case float32, float64:
				if !fieldValue.IsZero() {
					continue
				}

				if f := jsonValue.Float(); f != 0 {
					fieldValue.SetFloat(f)
				} else if f = txt.Float64(jsonValue.String()); f != 0 {
					fieldValue.SetFloat(f)
				}
			case uint, uint64:
				if !fieldValue.IsZero() {
					continue
				}

				if uintVal := jsonValue.Uint(); uintVal > 0 {
					fieldValue.SetUint(uintVal)
				} else if intVal, parseErr := strconv.ParseUint(strings.TrimSpace(jsonValue.String()), 10, 64); parseErr == nil && intVal > 0 {
					fieldValue.SetUint(intVal)
				}
			case []string:
				existing := fieldValue.Interface().([]string)
				fieldValue.Set(reflect.ValueOf(txt.AddToWords(existing, SanitizeUnicode(jsonValue.String()))))
			case Keywords:
				existing := fieldValue.Interface().(Keywords)
				fieldValue.Set(reflect.ValueOf(txt.AddToWords(existing, SanitizeUnicode(jsonValue.String()))))
			case projection.Type:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.Set(reflect.ValueOf(projection.Type(SanitizeUnicode(jsonValue.String()))))
			case string:
				if !fieldValue.IsZero() {
					continue
				}

				fieldValue.SetString(SanitizeUnicode(jsonValue.String()))
			case bool:
				if !fieldValue.IsZero() {
					continue
				}

				boolVal := false
				strVal := jsonValue.String()

				// Cast string to bool.
				switch strVal {
				case "1", "true":
					boolVal = true
				case "", "0", "false":
					boolVal = false
				default:
					boolVal = txt.NotEmpty(strVal)
				}

				fieldValue.SetBool(boolVal)
			default:
				log.Warnf("metadata: cannot assign value of type %s to %s (exiftool)", t, tagValue)
			}
		}
	}

	// Nanoseconds.
	if data.TakenNs <= 0 {
		for _, name := range exifSubSecTags {
			if s := data.json[name]; txt.IsPosInt(s) {
				data.TakenNs = txt.Int(s + strings.Repeat("0", 9-len(s)))
				break
			}
		}
	}

	// Set latitude and longitude if known and not already set.
	if data.Lat == 0 && data.Lng == 0 {
		if data.GPSPosition != "" {
			lat, lng := GpsToLatLng(data.GPSPosition)
			data.Lat, data.Lng = NormalizeGPS(lat, lng)
		} else if data.GPSLatitude != "" && data.GPSLongitude != "" {
			data.Lat, data.Lng = NormalizeGPS(GpsToDecimal(data.GPSLatitude), GpsToDecimal(data.GPSLongitude))
		}
	}

	if data.Altitude == 0 {
		// Parseable floating point number?
		if fl := GpsFloatRegexp.FindAllString(data.json["GPSAltitude"], -1); len(fl) != 1 {
			// Ignore.
		} else if alt, err := strconv.ParseFloat(fl[0], 64); err == nil && alt != 0 {
			data.Altitude = alt
		}
	}

	// Normalize capture time, local time, and time zone using the shared
	// resolver so the ExifTool JSON path and the XMP sidecar path produce
	// identical entity state for the same metadata.
	data.ResolveTimeZone(logName)

	// Use actual image width and height if available, see issue #2447.
	if jsonValues["ImageWidth"].Exists() && jsonValues["ImageHeight"].Exists() {
		if val := jsonValues["ImageWidth"].Int(); val > 0 {
			data.Width = int(val)
		}

		if val := jsonValues["ImageHeight"].Int(); val > 0 {
			data.Height = int(val)
		}
	}

	// Image orientation, see https://www.daveperrett.com/articles/2012/07/28/exif-orientation-handling-is-a-ghetto/.
	if orientation, ok := data.json["Orientation"]; ok && orientation != "" {
		switch orientation {
		case "1", "Horizontal (normal)":
			data.Orientation = 1
		case "2":
			data.Orientation = 2
		case "3", "Rotate 180 CW":
			data.Orientation = 3
		case "4":
			data.Orientation = 4
		case "5":
			data.Orientation = 5
		case "6", "Rotate 90 CW":
			data.Orientation = 6
		case "7":
			data.Orientation = 7
		case "8", "Rotate 270 CW":
			data.Orientation = 8
		}
	}

	if data.Orientation == 0 {
		// Set orientation based on rotation.
		switch data.Rotation {
		case 0:
			data.Orientation = 1
		case -180, 180:
			data.Orientation = 3
		case 90:
			data.Orientation = 6
		case -90, 270:
			data.Orientation = 8
		}
	}

	// Parse MWG-RS and Microsoft MP:RegionInfo face regions (people markers)
	// from the embedded XMP so the indexer can reconcile them onto markers.
	if faces := parseExiftoolFaces(j, FaceOptions{Orientation: data.Orientation, Width: data.Width, Height: data.Height}); len(faces) > 0 {
		data.Faces = faces
	}

	// Normalize codec name.
	data.Codec = strings.ToLower(data.Codec)
	if strings.Contains(data.Codec, CodecJpeg) { // JPEG Image?
		data.Codec = CodecJpeg
	} else if c, ok := video.Codecs[data.Codec]; ok { // Video codec?
		data.Codec = c
	} else if strings.HasPrefix(data.Codec, "a_") { // Audio codec?
		data.Codec = ""
	}

	// Validate and normalize optional DocumentID.
	if data.DocumentID != "" {
		data.DocumentID = rnd.SanitizeUUID(data.DocumentID)
	}

	// Validate and normalize optional InstanceID.
	if data.InstanceID != "" {
		data.InstanceID = rnd.SanitizeUUID(data.InstanceID)
	}

	// Promote GPano-style markers to equirectangular when ProjectionType is absent.
	if data.Projection == "" {
		if data.json["UsePanoramaViewer"] == "true" || data.json["IsPhotosphere"] == "true" {
			data.Projection = projection.Equirectangular.String()
		}
	}

	if projection.Equirectangular.Equal(data.Projection) {
		data.AddKeywords(KeywordPanorama)
	}

	if data.Caption != "" {
		data.AutoAddKeywords(data.Caption)
		data.Caption = SanitizeCaption(data.Caption)
	}

	data.Title = SanitizeTitle(data.Title)
	data.Subject = SanitizeMeta(data.Subject)
	data.Artist = SanitizeMeta(data.Artist)

	// Ignore numeric model names as they are probably invalid.
	if txt.IsUInt(data.LensModel) {
		data.LensModel = ""
	}

	// Flag Samsung/Google Motion Photos as live media.
	if data.HasVideoEmbedded && (data.MimeType == header.ContentTypeJpeg || data.MimeType == header.ContentTypeHeic) {
		data.MediaType = media.Live
	}

	return nil
}
