package meta

import (
	"encoding/xml"
	"os"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/txt"
)

// XmpDocument represents an XMP sidecar file.
type XmpDocument struct {
	XMLName xml.Name `xml:"xmpmeta" json:"xmpmeta,omitempty"`
	Text    string   `xml:",chardata" json:"text,omitempty"`
	X       string   `xml:"x,attr" json:"x,omitempty"`
	Xmptk   string   `xml:"xmptk,attr" json:"xmptk,omitempty"`
	RDF     struct {
		Text        string `xml:",chardata" json:"text,omitempty"`
		Rdf         string `xml:"rdf,attr" json:"rdf,omitempty"`
		Description struct {
			Text            string `xml:",chardata" json:"text,omitempty"`
			About           string `xml:"about,attr" json:"about,omitempty"`
			Xmp             string `xml:"xmp,attr" json:"xmp,omitempty"`
			Aux             string `xml:"aux,attr" json:"aux,omitempty"`
			ExifEX          string `xml:"exifEX,attr" json:"exifex,omitempty"`
			Photoshop       string `xml:"photoshop,attr" json:"photoshop,omitempty"`
			XmpMM           string `xml:"xmpMM,attr" json:"xmpmm,omitempty"`
			Dc              string `xml:"dc,attr" json:"dc,omitempty"`
			Tiff            string `xml:"tiff,attr" json:"tiff,omitempty"`
			Exif            string `xml:"exif,attr" json:"exif,omitempty"`
			XmpRights       string `xml:"xmpRights,attr" json:"xmprights,omitempty"`
			Iptc4xmpCore    string `xml:"Iptc4xmpCore,attr" json:"iptc4xmpcore,omitempty"`
			Iptc4xmpExt     string `xml:"Iptc4xmpExt,attr" json:"iptc4xmpext,omitempty"`
			CreatorTool     string `xml:"CreatorTool"`                                // ELE-L29 10.0.0.168(C431E2...
			ModifyDate      string `xml:"ModifyDate"`                                 // 2020-01-01T17:28:23.89961...
			CreateDate      string `xml:"CreateDate"`                                 // 2020-01-01T17:28:23
			MetadataDate    string `xml:"MetadataDate"`                               // 2020-01-01T17:28:23.89961...
			Rating          string `xml:"Rating"`                                     // 4
			FStopFavorite   string `xml:"http://www.fstopapp.com/xmp/ favorite,attr"` // 1
			Lens            string `xml:"Lens"`                                       // HUAWEI P30 Rear Main Came...
			LensModel       string `xml:"LensModel"`                                  // HUAWEI P30 Rear Main Came...
			DateCreated     string `xml:"DateCreated"`                                // 2020-01-01T17:28:25.72962...
			ColorMode       string `xml:"ColorMode"`                                  // 3
			ICCProfile      string `xml:"ICCProfile"`                                 // sRGB IEC61966-2.1
			AuthorsPosition string `xml:"AuthorsPosition"`                            // Maintainer
			DocumentID      string `xml:"DocumentID"`                                 // 2C678C1811D7095FD79CC822B...
			InstanceID      string `xml:"InstanceID"`                                 // 2C678C1811D7095FD79CC822B...
			Format          string `xml:"format"`                                     // image/jpeg
			Title           struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Alt  struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					Li   struct {
						Text string `xml:",chardata" json:"text,omitempty"` // Night Shift / Berlin / 20...
						Lang string `xml:"lang,attr" json:"lang,omitempty"`
					} `xml:"li" json:"li,omitempty"`
				} `xml:"Alt" json:"alt,omitempty"`
			} `xml:"title" json:"title,omitempty"`
			Creator struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Seq  struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					Li   string `xml:"li"` // Michael Mayer
				} `xml:"Seq" json:"seq,omitempty"`
			} `xml:"creator" json:"creator,omitempty"`
			Description struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Alt  struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					Li   struct {
						Text string `xml:",chardata" json:"text,omitempty"` // Example file for developm...
						Lang string `xml:"lang,attr" json:"lang,omitempty"`
					} `xml:"li" json:"li,omitempty"`
				} `xml:"Alt" json:"alt,omitempty"`
			} `xml:"description" json:"description,omitempty"`
			Subject struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Bag  struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					Li   []string `xml:"li"` // desk, coffee, computer
				} `xml:"Bag" json:"bag,omitempty"`
				Seq struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					Li   []string `xml:"li"` // desk, coffee, computer
				} `xml:"Seq" json:"seq,omitempty"`
			} `xml:"subject" json:"subject,omitempty"`
			Rights struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Alt  struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					Li   struct {
						Text string `xml:",chardata" json:"text,omitempty"` // This is an (edited) legal...
						Lang string `xml:"lang,attr" json:"lang,omitempty"`
					} `xml:"li" json:"li,omitempty"`
				} `xml:"Alt" json:"alt,omitempty"`
			} `xml:"rights" json:"rights,omitempty"`
			ImageWidth    string `xml:"ImageWidth"`  // 3648
			ImageLength   string `xml:"ImageLength"` // 2736
			BitsPerSample struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Seq  struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					Li   []string `xml:"li"` // 8
				} `xml:"Seq" json:"seq,omitempty"`
			} `xml:"BitsPerSample" json:"bitspersample,omitempty"`
			PhotometricInterpretation string `xml:"PhotometricInterpretation"` // 2
			Orientation               string `xml:"Orientation"`               // 0
			SamplesPerPixel           string `xml:"SamplesPerPixel"`           // 3
			YCbCrPositioning          string `xml:"YCbCrPositioning"`          // 1
			XResolution               string `xml:"XResolution"`               // 72/1
			YResolution               string `xml:"YResolution"`               // 72/1
			ResolutionUnit            string `xml:"ResolutionUnit"`            // 2
			Make                      string `xml:"Make"`                      // HUAWEI
			Model                     string `xml:"Model"`                     // ELE-L29
			ExifVersion               string `xml:"ExifVersion"`               // 0210
			FlashpixVersion           string `xml:"FlashpixVersion"`           // 0100
			ColorSpace                string `xml:"ColorProfile"`              // 1
			ComponentsConfiguration   struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Seq  struct {
					Text string   `xml:",chardata" json:"text,omitempty"`
					Li   []string `xml:"li"` // 1, 2, 3, 0
				} `xml:"Seq" json:"seq,omitempty"`
			} `xml:"ComponentsConfiguration" json:"componentsconfiguration,omitempty"`
			CompressedBitsPerPixel string `xml:"CompressedBitsPerPixel"` // 95/100
			PixelXDimension        string `xml:"PixelXDimension"`        // 3648
			PixelYDimension        string `xml:"PixelYDimension"`        // 2736
			DateTimeOriginal       string `xml:"DateTimeOriginal"`       // 2020-01-01T17:28:23
			ExposureTime           string `xml:"ExposureTime"`           // 20000000/1000000000
			FNumber                string `xml:"FNumber"`                // 180/100
			ExposureProgram        string `xml:"ExposureProgram"`        // 2
			ISOSpeedRatings        struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Seq  struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					Li   string `xml:"li"` // 200
				} `xml:"Seq" json:"seq,omitempty"`
			} `xml:"ISOSpeedRatings" json:"isospeedratings,omitempty"`
			ShutterSpeedValue string `xml:"ShutterSpeedValue"` // 298973/10000
			ApertureValue     string `xml:"ApertureValue"`     // 1695994/1000000
			BrightnessValue   string `xml:"BrightnessValue"`   // 0/1
			ExposureBiasValue string `xml:"ExposureBiasValue"` // 0/10
			MaxApertureValue  string `xml:"MaxApertureValue"`  // 169/100
			MeteringMode      string `xml:"MeteringMode"`      // 5
			LightSource       string `xml:"LightSource"`       // 1
			Flash             struct {
				Text       string `xml:",chardata" json:"text,omitempty"`
				ParseType  string `xml:"parseType,attr" json:"parsetype,omitempty"`
				Fired      string `xml:"Fired"`      // False
				Return     string `xml:"Return"`     // 0
				Mode       string `xml:"Mode"`       // 0
				Function   string `xml:"Function"`   // False
				RedEyeMode string `xml:"RedEyeMode"` // False
			} `xml:"Flash" json:"flash,omitempty"`
			FocalLength           string `xml:"FocalLength"`           // 5580/1000
			SensingMethod         string `xml:"SensingMethod"`         // 2
			FileSource            string `xml:"FileSource"`            // 3
			SceneType             string `xml:"SceneType"`             // 1
			CustomRendered        string `xml:"CustomRendered"`        // 1
			ExposureMode          string `xml:"ExposureMode"`          // 0
			WhiteBalance          string `xml:"WhiteBalance"`          // 0
			DigitalZoomRatio      string `xml:"DigitalZoomRatio"`      // 100/100
			FocalLengthIn35mmFilm string `xml:"FocalLengthIn35mmFilm"` // 27
			SceneCaptureType      string `xml:"SceneCaptureType"`      // 0
			GainControl           string `xml:"GainControl"`           // 0
			Contrast              string `xml:"Contrast"`              // 0
			Saturation            string `xml:"Saturation"`            // 0
			Sharpness             string `xml:"Sharpness"`             // 0
			SubjectDistanceRange  string `xml:"SubjectDistanceRange"`  // 0
			SubSecTime            string `xml:"SubSecTime"`            // 899614
			SubSecTimeOriginal    string `xml:"SubSecTimeOriginal"`    // 899614
			SubSecTimeDigitized   string `xml:"SubSecTimeDigitized"`   // 899614
			GPSVersionID          string `xml:"GPSVersionID"`          // 2.2.0.0
			GPSLatitude           string `xml:"GPSLatitude"`           // 52,27.5814N
			GPSLongitude          string `xml:"GPSLongitude"`          // 13,19.3099E
			GPSAltitudeRef        string `xml:"GPSAltitudeRef"`        // 1
			GPSAltitude           string `xml:"GPSAltitude"`           // 0/100
			GPSTimeStamp          string `xml:"GPSTimeStamp"`          // 2020-01-01T16:28:22Z
			Marked                string `xml:"Marked"`                // False
			WebStatement          string `xml:"WebStatement"`          // http://docs.photoprism.or...
			CreatorContactInfo    struct {
				Text        string `xml:",chardata" json:"text,omitempty"`
				ParseType   string `xml:"parseType,attr" json:"parsetype,omitempty"`
				CiAdrExtadr string `xml:"CiAdrExtadr"` // Zimmermannstr. 37
				CiAdrCity   string `xml:"CiAdrCity"`   // Berlin
				CiAdrPcode  string `xml:"CiAdrPcode"`  // 12163
				CiAdrCtry   string `xml:"CiAdrCtry"`   // Germany
				CiTelWork   string `xml:"CiTelWork"`   // +49123456789
				CiEmailWork string `xml:"CiEmailWork"` // hello@photoprism.org
				CiUrlWork   string `xml:"CiUrlWork"`   // https://photoprism.org/
			} `xml:"CreatorContactInfo" json:"creatorcontactinfo,omitempty"`
			PersonInImage struct {
				Text string `xml:",chardata" json:"text,omitempty"`
				Bag  struct {
					Text string `xml:",chardata" json:"text,omitempty"`
					Li   string `xml:"li"` // Gopher
				} `xml:"Bag" json:"bag,omitempty"`
			} `xml:"PersonInImage" json:"personinimage,omitempty"`
		} `xml:"Description" json:"description,omitempty"`
	} `xml:"RDF" json:"rdf,omitempty"`
}

// Load parses an XMP file and populates document values with its contents.
func (doc *XmpDocument) Load(filename string) error {
	data, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	return xml.Unmarshal(data, doc)
}

// Title returns the XMP document title.
func (doc *XmpDocument) Title() string {
	t := doc.RDF.Description.Title.Alt.Li.Text
	t2 := doc.RDF.Description.Title.Text
	if t != "" {
		return SanitizeTitle(t)
	} else if t2 != "" {
		return SanitizeTitle(t2)
	}
	return ""
}

// Artist returns the XMP document artist.
func (doc *XmpDocument) Artist() string {
	return SanitizeString(doc.RDF.Description.Creator.Seq.Li)
}

// Description returns the XMP document description.
func (doc *XmpDocument) Description() string {
	d := doc.RDF.Description.Description.Alt.Li.Text
	d2 := doc.RDF.Description.Description.Text
	if d != "" {
		return SanitizeDescription(d)
	} else if d2 != "" {
		return SanitizeTitle(d2)
	}
	return ""
}

// Copyright returns the XMP document copyright info.
func (doc *XmpDocument) Copyright() string {
	return SanitizeString(doc.RDF.Description.Rights.Alt.Li.Text)
}

// CameraMake returns the XMP document camera make name.
func (doc *XmpDocument) CameraMake() string {
	return SanitizeString(doc.RDF.Description.Make)
}

// CameraModel returns the XMP document camera model name.
func (doc *XmpDocument) CameraModel() string {
	return SanitizeString(doc.RDF.Description.Model)
}

// LensModel returns the XMP document lens model name.
func (doc *XmpDocument) LensModel() string {
	return SanitizeString(doc.RDF.Description.LensModel)
}

// TakenAt returns the XMP document taken date.
func (doc *XmpDocument) TakenAt(timeZone string) time.Time {
	taken := time.Time{} // Unknown

	s := SanitizeString(doc.RDF.Description.DateCreated)

	if s == "" {
		return taken
	}

	if dateTime := txt.DateTime(s, timeZone); !dateTime.IsZero() {
		return dateTime
	}

	return taken
}

// Keywords returns the XMP document keywords.
func (doc *XmpDocument) Keywords() string {
	s := doc.RDF.Description.Subject.Seq.Li

	return strings.Join(s, ", ")
}

// Favorite returns a favorite status in the XMP document.
func (doc *XmpDocument) Favorite() bool {
	fstop := doc.RDF.Description.FStopFavorite
	if fstop == "1" {
		return true
	}
	return false
}
