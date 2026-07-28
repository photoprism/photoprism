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

// jsonFaceName resolves the person name for a single ExifTool face region from
// its already index-aligned per-region values, falling back to a unique global
// PersonInImage entry referenced via rdfs:seeAlso.
func jsonFaceName(j gjson.Result, name, title, extension, reference string) string {
	switch {
	case name != "":
		return name
	case title != "":
		return title
	case extension != "":
		return extension
	}

	if reference != "Iptc4xmpExt:PersonInImage" {
		return ""
	}

	people := j.Get("PersonInImage").Array()
	unique := make([]string, 0, len(people))
	for _, result := range people {
		if v := strings.TrimSpace(result.String()); v != "" {
			unique = append(unique, v)
		}
	}
	if unique = txt.UniqueNames(unique); len(unique) == 1 {
		return unique[0]
	}

	return ""
}

// jsonArrayFit classifies an optional per-region ExifTool array against a region
// count of n. ExifTool omits absent struct members instead of padding them, so
// only len == n is safe to zip; an empty array means the member is absent for
// every region, and any other length means positional indexing would misassign.
func jsonArrayFit(arr []gjson.Result, n int) (aligned, sparse bool) {
	switch {
	case len(arr) == n:
		return true, false
	case len(arr) == 0:
		return false, false
	default:
		return false, true
	}
}

// parseExiftoolFaces extracts supported XMP face regions from ExifTool JSON.
// FaceRegions.Partial reports that ExifTool compacted a non-parallel set, such
// as mixed circle and rectangle shapes, so some regions could not be resolved.
func parseExiftoolFaces(j gjson.Result, options FaceOptions) FaceRegions {
	var faces []Face
	var tally regionTally

	partial := false
	// MWG-RS regions (XMP-mwg-rs): index-parallel arrays keyed by region index,
	// with absent optional members compacted out of their arrays.
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

	// Every region carries a center, so RegionAreaX and RegionAreaY are always
	// index-parallel and define how many regions exist.
	n := len(xs)
	if len(ys) < n {
		n = len(ys)
	}

	// recordFit reports whether arr can be indexed positionally, flagging the
	// parse partial when ExifTool compacted a sparse optional member.
	recordFit := func(arr []gjson.Result, count int) bool {
		aligned, sparse := jsonArrayFit(arr, count)
		if sparse {
			partial = true
		}
		return aligned
	}

	typesAligned := recordFit(types, n)
	namesAligned := recordFit(names, n)
	titlesAligned := recordFit(titles, n)
	extensionsAligned := recordFit(extensions, n)
	referencesAligned := recordFit(references, n)
	unitsAligned := recordFit(units, n)
	rotationsAligned := recordFit(rotations, n)

	// Shape resolves only when every region is a rectangle (parallel w and h) or
	// every region is a circle (parallel d); a mixed set leaves the shape arrays
	// compacted and unassignable, so it is reported partial rather than mis-sized.
	wAligned := recordFit(ws, n)
	hAligned := recordFit(hs, n)
	dAligned := recordFit(ds, n)
	rectMode := wAligned && hAligned
	circleMode := dAligned && !rectMode

	for i := 0; i < n; i++ {
		// Filter by type only when the array is index-aligned: mwg-rs:Type is
		// optional, so an entirely absent one is accepted as a face, while a
		// sparse one was flagged partial above and is skipped, not mislabeled.
		if typesAligned && !strings.EqualFold(types[i].String(), "Face") {
			continue
		} else if !typesAligned && len(types) > 0 {
			continue
		}

		tally.declare()

		var w, h, d float32
		switch {
		case rectMode:
			w, h = float32(ws[i].Float()), float32(hs[i].Float())
		case circleMode:
			d = float32(ds[i].Float())
		default:
			// Shape unresolvable for this set (already flagged partial above).
			continue
		}

		name := ""
		if namesAligned || titlesAligned || extensionsAligned || referencesAligned {
			nm, ti, ex, rf := "", "", "", ""
			if namesAligned {
				nm = names[i].String()
			}
			if titlesAligned {
				ti = titles[i].String()
			}
			if extensionsAligned {
				ex = extensions[i].String()
			}
			if referencesAligned {
				rf = references[i].String()
			}
			name = jsonFaceName(j, nm, ti, ex, rf)
		}

		unit := ""
		if unitsAligned {
			unit = units[i].String()
		}
		var rotation float32
		if rotationsAligned {
			rotation = float32(rotations[i].Float())
		}

		if f, ok := normalizeRegionMWG(
			name,
			float32(xs[i].Float()), float32(ys[i].Float()),
			w, h, d, unit, rotation,
			appliedW, appliedH, options.Orientation,
		); ok {
			faces = append(faces, f)
			tally.resolve()
		}
	}

	// Microsoft MP:RegionInfo (XMP-MP): a self-contained rectangle string per
	// region plus an optional display name.
	mpNames := j.Get("RegionPersonDisplayName").Array()
	mpRects := j.Get("RegionRectangle").Array()
	mpNamesAligned := recordFit(mpNames, len(mpRects))

	for i := 0; i < len(mpRects); i++ {
		rect := mpRects[i].String()
		if rect == "" {
			continue
		}

		tally.declare()

		name := ""
		if mpNamesAligned {
			name = mpNames[i].String()
		}
		if f, ok := normalizeRegionMP(name, rect, options.Orientation); ok {
			faces = append(faces, f)
			tally.resolve()
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
	dlyUnits := j.Get("ACDSeeRegionDLYAreaUnit").Array()
	algXs := j.Get("ACDSeeRegionALGAreaX").Array()
	algYs := j.Get("ACDSeeRegionALGAreaY").Array()
	algWs := j.Get("ACDSeeRegionALGAreaW").Array()
	algHs := j.Get("ACDSeeRegionALGAreaH").Array()
	algUnits := j.Get("ACDSeeRegionALGAreaUnit").Array()

	acdW := int(j.Get("ACDSeeRegionAppliedToDimensionsW").Int())
	acdH := int(j.Get("ACDSeeRegionAppliedToDimensionsH").Int())
	if acdW <= 0 || acdH <= 0 {
		acdW, acdH = options.Width, options.Height
	}

	// ACDSeeRegionType is optional, so the region count comes from the coordinate
	// arrays as well: deriving it from the type array alone would silently report
	// zero regions for a writer that omits the type.
	nAcd := max(len(acdTypes), len(dlyXs), len(algXs))
	acdTypesAligned := recordFit(acdTypes, nAcd)
	acdNamesAligned := recordFit(acdNames, nAcd)

	// Evaluate every recordFit before combining: && short-circuits and would skip
	// the partial bookkeeping for the remaining arrays.
	dlyXAligned, dlyYAligned := recordFit(dlyXs, nAcd), recordFit(dlyYs, nAcd)
	dlyWAligned, dlyHAligned := recordFit(dlyWs, nAcd), recordFit(dlyHs, nAcd)
	algXAligned, algYAligned := recordFit(algXs, nAcd), recordFit(algYs, nAcd)
	algWAligned, algHAligned := recordFit(algWs, nAcd), recordFit(algHs, nAcd)
	dlyAligned := dlyXAligned && dlyYAligned && dlyWAligned && dlyHAligned
	algAligned := algXAligned && algYAligned && algWAligned && algHAligned
	dlyUnitsAligned := recordFit(dlyUnits, nAcd)
	algUnitsAligned := recordFit(algUnits, nAcd)

	for i := 0; i < nAcd; i++ {
		if acdTypesAligned && !strings.EqualFold(acdTypes[i].String(), "Face") {
			continue
		} else if !acdTypesAligned && len(acdTypes) > 0 {
			continue
		}

		tally.declare()

		var xr, yr, wr, hr gjson.Result
		unit := ""
		switch {
		case dlyAligned:
			xr, yr, wr, hr = dlyXs[i], dlyYs[i], dlyWs[i], dlyHs[i]
			if dlyUnitsAligned {
				unit = dlyUnits[i].String()
			}
		case algAligned:
			xr, yr, wr, hr = algXs[i], algYs[i], algWs[i], algHs[i]
			if algUnitsAligned {
				unit = algUnits[i].String()
			}
		default:
			partial = true
			continue
		}

		name := ""
		if acdNamesAligned {
			name = acdNames[i].String()
		}
		if f, ok := normalizeRegionMWG(
			name,
			float32(xr.Float()), float32(yr.Float()),
			float32(wr.Float()), float32(hr.Float()), 0,
			unit, 0, acdW, acdH, options.Orientation,
		); ok {
			faces = append(faces, f)
			tally.resolve()
		}
	}

	// ExifTool flattens the region containers away, so an empty region list is
	// indistinguishable from a file that carries none. Declaring only a non-empty
	// set keeps markers from being deleted on evidence this projection lacks.
	return FaceRegions{
		Faces:    DedupFaces(faces),
		Declared: n > 0 || len(mpRects) > 0 || nAcd > 0,
		Partial:  partial || tally.partial(),
	}
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
	if regions := parseExiftoolFaces(j, FaceOptions{Orientation: data.Orientation, Width: data.Width, Height: data.Height}); regions.Declared || len(regions.Faces) > 0 {
		data.Faces = regions.Faces
		data.FacesDeclared = regions.Declared
		data.FacesPartial = regions.Partial
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
