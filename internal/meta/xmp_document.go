package meta

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// xmpMaxFileSize caps XMP sidecar size to mitigate denial-of-service via
// pathologically large files. Real-world sidecars range from ~5 KB to
// ~200 KB; 1 MiB leaves generous headroom.
const xmpMaxFileSize = 1 * 1024 * 1024

// xmpMaxDepth caps element nesting to mitigate depth-bomb attacks.
// Natural XMP nests to ~10–12 levels; 64 is well above any legitimate use.
const xmpMaxDepth = 64

// ErrXmpFileTooLarge is returned when an XMP sidecar exceeds xmpMaxFileSize.
var ErrXmpFileTooLarge = errors.New("xmp: file size exceeds limit")

// ErrXmpTooDeep is returned when an XMP document nests deeper than xmpMaxDepth.
var ErrXmpTooDeep = errors.New("xmp: element nesting exceeds limit")

// xmpNamespaces binds canonical XMP namespace prefixes to their URIs.
// Adding a prefix here makes it available to every XPath expression
// compiled with mustCompile.
var xmpNamespaces = map[string]string{
	"xml":          "http://www.w3.org/XML/1998/namespace",
	"rdf":          "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"x":            "adobe:ns:meta/",
	"xmp":          "http://ns.adobe.com/xap/1.0/",
	"xmpMM":        "http://ns.adobe.com/xap/1.0/mm/",
	"xmpRights":    "http://ns.adobe.com/xap/1.0/rights/",
	"xmpDM":        "http://ns.adobe.com/xmp/1.0/DynamicMedia/",
	"dc":           "http://purl.org/dc/elements/1.1/",
	"tiff":         "http://ns.adobe.com/tiff/1.0/",
	"exif":         "http://ns.adobe.com/exif/1.0/",
	"exifEX":       "http://cipa.jp/exif/1.0/",
	"aux":          "http://ns.adobe.com/exif/1.0/aux/",
	"photoshop":    "http://ns.adobe.com/photoshop/1.0/",
	"GPano":        "http://ns.google.com/photos/1.0/panorama/",
	"Iptc4xmpCore": "http://iptc.org/std/Iptc4xmpCore/1.0/xmlns/",
	"Iptc4xmpExt":  "http://iptc.org/std/Iptc4xmpExt/2008-02-29/",
	"lr":           "http://ns.adobe.com/lightroom/1.0/",
	"fstop":        "http://www.fstopapp.com/xmp/",
}

// chainXPath is an ordered list of pre-compiled XPath expressions
// evaluated left-to-right; the first non-empty match wins.
type chainXPath []*xpath.Expr

// firstNonEmpty returns the trimmed inner text of the first matching
// node; empty string when no link in the chain produces a value.
func (c chainXPath) firstNonEmpty(root *xmlquery.Node) string {
	if root == nil {
		return ""
	}
	for _, expr := range c {
		if n := xmlquery.QuerySelector(root, expr); n != nil {
			if s := strings.TrimSpace(n.InnerText()); s != "" {
				return s
			}
		}
	}
	return ""
}

// queryAll returns the trimmed text of every matching node in document
// order; empty matches are dropped.
func queryAll(root *xmlquery.Node, expr *xpath.Expr) []string {
	if root == nil || expr == nil {
		return nil
	}
	nodes := xmlquery.QuerySelectorAll(root, expr)
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if s := strings.TrimSpace(n.InnerText()); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// mustCompile compiles an XPath expression bound to xmpNamespaces and
// panics on a malformed expression. Intended for package-init wiring.
func mustCompile(expr string) *xpath.Expr {
	compiled, err := xpath.CompileWithNS(expr, xmpNamespaces)
	if err != nil {
		panic(fmt.Sprintf("xmp: invalid XPath %q: %v", expr, err))
	}
	return compiled
}

// elemOrAttr matches the named element or the same property expressed
// as an attribute on rdf:Description (digiKam emits attributes, Adobe
// emits child elements; both are valid RDF/XML).
func elemOrAttr(qname string) *xpath.Expr {
	return mustCompile(fmt.Sprintf("//%s | //rdf:Description/@%s", qname, qname))
}

// Pre-compiled query handles. Compile-once-at-init amortizes XPath
// parsing across every sidecar in an indexer run.
var (
	// xmpTitleChain: dc:title (Alt/x-default) → first li → bare text → photoshop:Headline.
	xmpTitleChain = chainXPath{
		mustCompile("//dc:title/rdf:Alt/rdf:li[@xml:lang='x-default']"),
		mustCompile("(//dc:title/rdf:Alt/rdf:li)[1]"),
		mustCompile("//dc:title[not(rdf:Alt)]"),
		mustCompile("//photoshop:Headline"),
	}

	// xmpDescriptionChain: dc:description Alt-language fallback.
	xmpDescriptionChain = chainXPath{
		mustCompile("//dc:description/rdf:Alt/rdf:li[@xml:lang='x-default']"),
		mustCompile("(//dc:description/rdf:Alt/rdf:li)[1]"),
		mustCompile("//dc:description[not(rdf:Alt)]"),
	}

	// xmpRightsChain: dc:rights Alt-language → xmpRights:WebStatement.
	// WebStatement is a URL, not free text, but the embedded path uses
	// the same fallback.
	xmpRightsChain = chainXPath{
		mustCompile("//dc:rights/rdf:Alt/rdf:li[@xml:lang='x-default']"),
		mustCompile("(//dc:rights/rdf:Alt/rdf:li)[1]"),
		mustCompile("//dc:rights[not(rdf:Alt)]"),
		mustCompile("//xmpRights:WebStatement"),
	}

	// xmpLicenseChain: xmpRights:UsageTerms (lang-alt).
	xmpLicenseChain = chainXPath{
		mustCompile("//xmpRights:UsageTerms/rdf:Alt/rdf:li[@xml:lang='x-default']"),
		mustCompile("(//xmpRights:UsageTerms/rdf:Alt/rdf:li)[1]"),
		mustCompile("//xmpRights:UsageTerms[not(rdf:Alt)]"),
	}

	// xmpSoftwareChain: xmp:CreatorTool.
	xmpSoftwareChain = chainXPath{elemOrAttr("xmp:CreatorTool")}

	// xmpDocumentIDChain: xmpMM:OriginalDocumentID (asset-stable) →
	// xmpMM:DocumentID (per-derivative) → dc:identifier (legacy).
	xmpDocumentIDChain = chainXPath{
		elemOrAttr("xmpMM:OriginalDocumentID"),
		elemOrAttr("xmpMM:DocumentID"),
		elemOrAttr("dc:identifier"),
	}

	// xmpInstanceIDChain: xmpMM:InstanceID.
	xmpInstanceIDChain = chainXPath{elemOrAttr("xmpMM:InstanceID")}

	// xmpArtistChain: first dc:creator/rdf:Seq entry.
	xmpArtistChain = chainXPath{mustCompile("(//dc:creator/rdf:Seq/rdf:li)[1]")}

	// xmpCameraMakeChain: tiff:Make.
	xmpCameraMakeChain = chainXPath{elemOrAttr("tiff:Make")}

	// xmpCameraModelChain: tiff:Model. No XMP fallback exists; aux
	// and tiff UniqueCameraModel are not defined XMP properties.
	xmpCameraModelChain = chainXPath{elemOrAttr("tiff:Model")}

	// xmpLensModelChain: exifEX:LensModel (modern) → aux:Lens →
	// aux:LensID. aux is retained because Lightroom and Camera Raw
	// still emit it.
	xmpLensModelChain = chainXPath{
		elemOrAttr("exifEX:LensModel"),
		elemOrAttr("aux:Lens"),
		elemOrAttr("aux:LensID"),
	}

	// xmpTakenAtChain: photoshop:DateCreated → exif:DateTimeOriginal →
	// xmp:CreateDate. SubSecTimeOriginal is joined in TakenAt() when
	// the parsed datetime carries no fractional component.
	xmpTakenAtChain = chainXPath{
		elemOrAttr("photoshop:DateCreated"),
		elemOrAttr("exif:DateTimeOriginal"),
		elemOrAttr("xmp:CreateDate"),
	}

	// xmpSubSecTimeChain: exif:SubSecTimeOriginal — fractional seconds
	// of the capture time as a digit string.
	xmpSubSecTimeChain = chainXPath{elemOrAttr("exif:SubSecTimeOriginal")}

	// xmpTimeOffsetChain: EXIF 2.31 cascade.
	xmpTimeOffsetChain = chainXPath{
		elemOrAttr("exif:OffsetTimeOriginal"),
		elemOrAttr("exif:OffsetTime"),
		elemOrAttr("exif:OffsetTimeDigitized"),
	}

	// xmpCreatedAtChain: xmp:CreateDate → xmpDM:CreationDate (video).
	xmpCreatedAtChain = chainXPath{
		elemOrAttr("xmp:CreateDate"),
		elemOrAttr("xmpDM:CreationDate"),
	}

	// xmpCameraSerialChain: exifEX:SerialNumber (= EXIF
	// BodySerialNumber) → aux:SerialNumber.
	xmpCameraSerialChain = chainXPath{
		elemOrAttr("exifEX:SerialNumber"),
		elemOrAttr("aux:SerialNumber"),
	}

	// xmpLensMakeChain: exifEX:LensMake.
	xmpLensMakeChain = chainXPath{elemOrAttr("exifEX:LensMake")}

	// xmpCameraOwnerChain: aux:OwnerName.
	xmpCameraOwnerChain = chainXPath{elemOrAttr("aux:OwnerName")}

	// xmpProjectionChain: GPano:ProjectionType.
	xmpProjectionChain = chainXPath{elemOrAttr("GPano:ProjectionType")}

	// xmpColorProfileChain: photoshop:ICCProfile.
	xmpColorProfileChain = chainXPath{elemOrAttr("photoshop:ICCProfile")}

	// xmpApertureChain: exif:ApertureValue (APEX).
	xmpApertureChain = chainXPath{elemOrAttr("exif:ApertureValue")}

	// xmpFNumberChain: exif:FNumber.
	xmpFNumberChain = chainXPath{elemOrAttr("exif:FNumber")}

	// xmpFocalLengthChain: exif:FocalLength → exif:FocalLengthIn35mmFilm.
	xmpFocalLengthChain = chainXPath{
		elemOrAttr("exif:FocalLength"),
		elemOrAttr("exif:FocalLengthIn35mmFilm"),
	}

	// xmpIsoChain: exifEX:PhotographicSensitivity (modern) →
	// exif:ISOSpeedRatings/rdf:Seq[1] (deprecated but widely emitted).
	xmpIsoChain = chainXPath{
		elemOrAttr("exifEX:PhotographicSensitivity"),
		mustCompile("(//exif:ISOSpeedRatings/rdf:Seq/rdf:li)[1]"),
		mustCompile("//rdf:Description/@exif:ISOSpeedRatings"),
	}

	// xmpExposureTimeChain: exif:ExposureTime as rational seconds.
	xmpExposureTimeChain = chainXPath{elemOrAttr("exif:ExposureTime")}

	// xmpShutterSpeedChain: exif:ShutterSpeedValue (APEX-encoded);
	// converted to seconds in Exposure() via t = 2^(-Tv).
	xmpShutterSpeedChain = chainXPath{elemOrAttr("exif:ShutterSpeedValue")}

	// xmpFlashFiredChain: only the Fired sub-field; other Flash
	// sub-fields are intentionally ignored because data.Flash is bool.
	xmpFlashFiredChain = chainXPath{
		mustCompile("//exif:Flash/exif:Fired"),
		mustCompile("//exif:Flash/@exif:Fired"),
		mustCompile("//exif:Flash/rdf:Description/@exif:Fired"),
	}

	// xmpNotesChain: exif:UserComment (lang-alt).
	xmpNotesChain = chainXPath{
		mustCompile("//exif:UserComment/rdf:Alt/rdf:li[@xml:lang='x-default']"),
		mustCompile("(//exif:UserComment/rdf:Alt/rdf:li)[1]"),
		mustCompile("//exif:UserComment[not(rdf:Alt)]"),
	}

	// xmpSubjectBag / xmpSubjectSeq: dc:subject containers. dc:subject is
	// Adobe's "Keywords" panel, but PhotoPrism maps it to the descriptive
	// Details.Subject field, not the keyword list. Adobe, Darktable and
	// digiKam emit Bag; Apple Photos and the previous reader emit Seq.
	xmpSubjectBag = mustCompile("//dc:subject/rdf:Bag/rdf:li")
	xmpSubjectSeq = mustCompile("//dc:subject/rdf:Seq/rdf:li")

	// xmpPersonBag / xmpPersonSeq: Iptc4xmpExt:PersonInImage containers
	// (names of people depicted) — a Subject cascade fallback.
	xmpPersonBag = mustCompile("//Iptc4xmpExt:PersonInImage/rdf:Bag/rdf:li")
	xmpPersonSeq = mustCompile("//Iptc4xmpExt:PersonInImage/rdf:Seq/rdf:li")

	// xmpHierarchicalBag / xmpHierarchicalSeq: lr:hierarchicalSubject
	// containers (Lightroom "Nature|Animals" paths) — a Subject fallback.
	xmpHierarchicalBag = mustCompile("//lr:hierarchicalSubject/rdf:Bag/rdf:li")
	xmpHierarchicalSeq = mustCompile("//lr:hierarchicalSubject/rdf:Seq/rdf:li")

	// xmpFavoriteAttr: F-Stop favorite attribute on rdf:Description.
	xmpFavoriteAttr = mustCompile("//rdf:Description/@fstop:favorite")

	// xmpGPSLatitudeChain reads exif:GPSLatitude; sign composition
	// with GPSLatitudeRef happens in Lat().
	xmpGPSLatitudeChain    = chainXPath{elemOrAttr("exif:GPSLatitude")}
	xmpGPSLatitudeRefChain = chainXPath{elemOrAttr("exif:GPSLatitudeRef")}

	// xmpGPSLongitudeChain: analogous to latitude.
	xmpGPSLongitudeChain    = chainXPath{elemOrAttr("exif:GPSLongitude")}
	xmpGPSLongitudeRefChain = chainXPath{elemOrAttr("exif:GPSLongitudeRef")}

	// xmpGPSAltitudeChain: rationals like "3450/100" parsed in Altitude().
	xmpGPSAltitudeChain    = chainXPath{elemOrAttr("exif:GPSAltitude")}
	xmpGPSAltitudeRefChain = chainXPath{elemOrAttr("exif:GPSAltitudeRef")}

	// xmpGPSTimeStampChain: combined ISO 8601 datetime (canonical XMP).
	xmpGPSTimeStampChain = chainXPath{elemOrAttr("exif:GPSTimeStamp")}

	// xmpGPSDateStampChain: legacy split-form fallback for TakenGps.
	xmpGPSDateStampChain = chainXPath{elemOrAttr("exif:GPSDateStamp")}
)

// XmpDocument represents a parsed XMP sidecar; populate via Load.
type XmpDocument struct {
	doc *xmlquery.Node
}

// Load reads an XMP sidecar and enforces the size and depth security
// limits. XXE/DTD attacks are mitigated by encoding/xml defaults
// (xmp_security_test.go guards). io.LimitReader caps Parse defensively
// in case the file grows between Stat and Read.
func (doc *XmpDocument) Load(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info.Size() > xmpMaxFileSize {
		return fmt.Errorf("%w: %s (%d bytes)", ErrXmpFileTooLarge, clean.Log(filename), info.Size())
	}

	f, err := os.Open(filename) //nolint:gosec // sidecar reading is the documented purpose
	if err != nil {
		return err
	}
	defer f.Close()

	parsed, err := xmlquery.Parse(io.LimitReader(f, xmpMaxFileSize))
	if err != nil {
		return fmt.Errorf("xmp: %s: %w", clean.Log(filename), err)
	}

	if d := maxDepth(parsed, 0); d > xmpMaxDepth {
		return fmt.Errorf("%w: %s (got %d)", ErrXmpTooDeep, clean.Log(filename), d)
	}

	doc.doc = parsed
	return nil
}

// maxDepth returns the deepest element-nesting level reachable from n.
// Non-element nodes (text, comment, attribute) do not inflate the count.
func maxDepth(n *xmlquery.Node, current int) int {
	if n == nil {
		return current
	}
	deepest := current
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != xmlquery.ElementNode {
			continue
		}
		if d := maxDepth(c, current+1); d > deepest {
			deepest = d
		}
	}
	return deepest
}

// Title returns the XMP document title.
// Priority: dc:title (Alt/x-default → first rdf:Alt entry → bare text) → photoshop:Headline.
func (doc *XmpDocument) Title() string {
	return SanitizeTitle(xmpTitleChain.firstNonEmpty(doc.doc))
}

// Description returns the caption / image description.
// Priority: dc:description (Alt/x-default → first rdf:Alt entry → bare text).
func (doc *XmpDocument) Description() string {
	return SanitizeCaption(xmpDescriptionChain.firstNonEmpty(doc.doc))
}

// Copyright returns the rights statement.
// Priority: dc:rights (Alt/x-default → first rdf:Alt entry → bare text) → xmpRights:WebStatement.
// WebStatement is a URL approximation of the rights text but matches the embedded path's behavior.
func (doc *XmpDocument) Copyright() string {
	return SanitizeString(xmpRightsChain.firstNonEmpty(doc.doc))
}

// Artist returns the first creator name.
// Priority: first entry of dc:creator/rdf:Seq.
func (doc *XmpDocument) Artist() string {
	return SanitizeString(xmpArtistChain.firstNonEmpty(doc.doc))
}

// CameraMake returns the camera manufacturer.
// Priority: tiff:Make (single-link, namespace-bound).
func (doc *XmpDocument) CameraMake() string {
	return SanitizeString(xmpCameraMakeChain.firstNonEmpty(doc.doc))
}

// CameraModel returns the camera model.
// Priority: tiff:Model (single-link; UniqueCameraModel is a DNG TIFF tag,
// not an XMP property).
func (doc *XmpDocument) CameraModel() string {
	return SanitizeString(xmpCameraModelChain.firstNonEmpty(doc.doc))
}

// LensModel returns the lens model.
// Priority: exifEX:LensModel (modern EXIF 2.3) → aux:Lens → aux:LensID (legacy Adobe).
func (doc *XmpDocument) LensModel() string {
	return SanitizeString(xmpLensModelChain.firstNonEmpty(doc.doc))
}

// TakenAt parses the capture timestamp.
// Priority: photoshop:DateCreated → exif:DateTimeOriginal → xmp:CreateDate.
// Composition: exif:SubSecTimeOriginal is joined when the primary date
// string lacks fractional seconds (never overrides an existing fraction).
// Diverges from the embedded path (DateTimeOriginal-first) because Adobe
// sidecars treat photoshop:DateCreated as authoritative; SrcXmp > SrcMeta
// at the entity layer resolves any cross-path disagreement.
func (doc *XmpDocument) TakenAt(timeZone string) time.Time {
	s := SanitizeString(xmpTakenAtChain.firstNonEmpty(doc.doc))
	if s == "" {
		return time.Time{}
	}
	t := txt.ParseTime(s, timeZone)
	if t.IsZero() || t.Nanosecond() != 0 {
		return t
	}
	if subSec := SanitizeString(xmpSubSecTimeChain.firstNonEmpty(doc.doc)); subSec != "" {
		if ns := parseSubSec(subSec); ns > 0 {
			t = t.Add(time.Duration(ns) * time.Nanosecond)
		}
	}
	return t
}

// TakenNs returns the nanosecond fraction of the capture timestamp.
// Priority: exif:SubSecTimeOriginal (single-link). Independent of TakenAt's
// join — callers that want only the sub-second component consume this.
func (doc *XmpDocument) TakenNs() int {
	subSec := SanitizeString(xmpSubSecTimeChain.firstNonEmpty(doc.doc))
	if subSec == "" {
		return 0
	}
	return parseSubSec(subSec)
}

// CreatedAt parses the file-creation timestamp (distinct from TakenAt:
// TakenAt = capture time, CreatedAt = first write of the digital file).
// Priority: xmp:CreateDate → xmpDM:CreationDate (Dynamic Media fallback).
func (doc *XmpDocument) CreatedAt(timeZone string) time.Time {
	s := SanitizeString(xmpCreatedAtChain.firstNonEmpty(doc.doc))
	if s == "" {
		return time.Time{}
	}
	return txt.ParseTime(s, timeZone)
}

// TimeOffset returns the timezone offset string ("+02:00").
// Priority: exif:OffsetTimeOriginal → exif:OffsetTime → exif:OffsetTimeDigitized (EXIF 2.31 cascade).
func (doc *XmpDocument) TimeOffset() string {
	return SanitizeString(xmpTimeOffsetChain.firstNonEmpty(doc.doc))
}

// CameraSerial returns the camera body serial number.
// Priority: exifEX:SerialNumber (= EXIF BodySerialNumber 0xA431) →
// aux:SerialNumber (legacy Adobe, still emitted by Lightroom).
func (doc *XmpDocument) CameraSerial() string {
	return SanitizeString(xmpCameraSerialChain.firstNonEmpty(doc.doc))
}

// LensMake returns the lens manufacturer.
// Priority: exifEX:LensMake (single-link; aux:LensMake is not defined).
func (doc *XmpDocument) LensMake() string {
	return SanitizeString(xmpLensMakeChain.firstNonEmpty(doc.doc))
}

// CameraOwner returns the camera-owner name.
// Priority: aux:OwnerName (single-link).
func (doc *XmpDocument) CameraOwner() string {
	return SanitizeString(xmpCameraOwnerChain.firstNonEmpty(doc.doc))
}

// Projection returns the panoramic projection type.
// Priority: GPano:ProjectionType (single-link). Google Photo Sphere
// defines only "equirectangular"; other values pass through unchanged.
func (doc *XmpDocument) Projection() string {
	return SanitizeString(xmpProjectionChain.firstNonEmpty(doc.doc))
}

// ColorProfile returns the embedded ICC profile description.
// Priority: photoshop:ICCProfile (single-link).
func (doc *XmpDocument) ColorProfile() string {
	return SanitizeString(xmpColorProfileChain.firstNonEmpty(doc.doc))
}

// Aperture returns the APEX-encoded aperture value.
// Priority: exif:ApertureValue (single-link, parsed as rational "180/100" → 1.8).
func (doc *XmpDocument) Aperture() float32 {
	return float32(rationalAccessor(xmpApertureChain, doc.doc))
}

// FNumber returns the f-number (focal length / entrance-pupil diameter, e.g. 1.8 for f/1.8).
// Priority: exif:FNumber (single-link, parsed as rational).
func (doc *XmpDocument) FNumber() float32 {
	return float32(rationalAccessor(xmpFNumberChain, doc.doc))
}

// FocalLength returns the focal length in millimeters (rounded; the
// data field is int and sub-mm precision is discarded).
// Priority: exif:FocalLength (native) → exif:FocalLengthIn35mmFilm.
func (doc *XmpDocument) FocalLength() int {
	return int(math.Round(rationalAccessor(xmpFocalLengthChain, doc.doc)))
}

// Iso returns the ISO sensitivity.
// Priority: exifEX:PhotographicSensitivity (EXIF 2.3) → first
// exif:ISOSpeedRatings/rdf:Seq entry (deprecated but widely emitted).
func (doc *XmpDocument) Iso() int {
	val := xmpIsoChain.firstNonEmpty(doc.doc)
	if n, err := strconv.Atoi(val); err == nil && n > 0 {
		return n
	}
	return 0
}

// Exposure returns the exposure time as "1/250", "0.5", "30" etc.
// Priority: exif:ExposureTime (rational seconds) → exif:ShutterSpeedValue
// (APEX-encoded, converted via t = 2^(-Tv)).
func (doc *XmpDocument) Exposure() string {
	if val := xmpExposureTimeChain.firstNonEmpty(doc.doc); val != "" {
		if secs, ok := parseRational(val); ok && secs > 0 {
			return formatExposure(secs)
		}
	}
	if val := xmpShutterSpeedChain.firstNonEmpty(doc.doc); val != "" {
		if apex, ok := parseRational(val); ok {
			return formatExposure(apexToSeconds(apex))
		}
	}
	return ""
}

// Flash reports whether the flash fired.
// Composition: only exif:Flash/Fired is read; other sub-fields
// (Function, Mode, Return, RedEyeMode) are ignored because data.Flash
// is a single boolean.
func (doc *XmpDocument) Flash() bool {
	return txt.Bool(xmpFlashFiredChain.firstNonEmpty(doc.doc))
}

// Notes returns the user-comment text.
// Priority: exif:UserComment (lang-alt: x-default → first rdf:Alt entry → bare text).
func (doc *XmpDocument) Notes() string {
	return SanitizeString(xmpNotesChain.firstNonEmpty(doc.doc))
}

// joinBagOrSeq joins an rdf:Bag container's entries (Adobe/Darktable/digiKam)
// or, when no Bag is present, the rdf:Seq entries (Apple, older writers) with
// ", ". Bag wins when both are present.
func (doc *XmpDocument) joinBagOrSeq(bag, seq *xpath.Expr) string {
	if v := queryAll(doc.doc, bag); len(v) > 0 {
		return strings.Join(v, ", ")
	}
	return strings.Join(queryAll(doc.doc, seq), ", ")
}

// Subject returns descriptive subject text for the Details.Subject field,
// matching the ExifTool Subject cascade so the XMP and embedded/ExifTool JSON
// paths fill meta.Data.Subject identically. Priority: dc:subject (Adobe's
// "Keywords" panel, present in virtually all tagged files) →
// Iptc4xmpExt:PersonInImage → lr:hierarchicalSubject. The first non-empty
// container wins, joined with ", "; entries keep their spaces.
//
// The PersonInImage and hierarchicalSubject fallbacks are interim sources for
// the free-text Subject field; advanced parsing will route them to dedicated
// meta.Data.Subjects (people) and meta.Data.Labels containers.
func (doc *XmpDocument) Subject() string {
	if v := doc.joinBagOrSeq(xmpSubjectBag, xmpSubjectSeq); v != "" {
		return v
	}
	if v := doc.joinBagOrSeq(xmpPersonBag, xmpPersonSeq); v != "" {
		return v
	}
	return doc.joinBagOrSeq(xmpHierarchicalBag, xmpHierarchicalSeq)
}

// Favorite reports the F-Stop custom-namespace favorite flag.
// Priority: rdf:Description/@fstop:favorite (single-link, "1" → true).
func (doc *XmpDocument) Favorite() bool {
	if doc.doc == nil {
		return false
	}
	if n := xmlquery.QuerySelector(doc.doc, xmpFavoriteAttr); n != nil {
		return strings.TrimSpace(n.InnerText()) == "1"
	}
	return false
}

// License returns the XMP license statement.
// Priority: xmpRights:UsageTerms (lang-alt: x-default → first rdf:Alt entry → bare text).
func (doc *XmpDocument) License() string {
	return SanitizeString(xmpLicenseChain.firstNonEmpty(doc.doc))
}

// Software returns the application that wrote the file.
// Priority: xmp:CreatorTool (single-link).
func (doc *XmpDocument) Software() string {
	return SanitizeString(xmpSoftwareChain.firstNonEmpty(doc.doc))
}

// DocumentID returns the XMP document identifier (asset-stable across
// derivatives; "xmp.did:" / "xmp.iid:" prefixes are preserved and
// stripped downstream when matching across embedded/sidecar paths).
// Priority: xmpMM:OriginalDocumentID → xmpMM:DocumentID → dc:identifier.
func (doc *XmpDocument) DocumentID() string {
	return SanitizeString(xmpDocumentIDChain.firstNonEmpty(doc.doc))
}

// InstanceID returns the XMP instance identifier (per-derivative).
// Priority: xmpMM:InstanceID (single-link; each saved derivative gets a fresh ID).
func (doc *XmpDocument) InstanceID() string {
	return SanitizeString(xmpInstanceIDChain.firstNonEmpty(doc.doc))
}

// Lat returns the GPS latitude as a decimal float.
// Priority: exif:GPSLatitude (single-link).
// Composition: cardinal in the value (Adobe form "52,30.4567N") wins;
// otherwise exif:GPSLatitudeRef supplies the sign. See gpsCoord.
func (doc *XmpDocument) Lat() float64 {
	return gpsCoord(xmpGPSLatitudeChain, xmpGPSLatitudeRefChain, doc.doc)
}

// Lng returns the GPS longitude as a decimal float.
// Priority: exif:GPSLongitude (single-link); composition mirrors Lat
// with E/W cardinal handling.
func (doc *XmpDocument) Lng() float64 {
	return gpsCoord(xmpGPSLongitudeChain, xmpGPSLongitudeRefChain, doc.doc)
}

// gpsCoord composes a single GPS coordinate, applying the cardinal
// from refChain when the value lacks an N/S/E/W suffix.
func gpsCoord(valueChain, refChain chainXPath, root *xmlquery.Node) float64 {
	val := valueChain.firstNonEmpty(root)
	if val == "" {
		return 0
	}
	decimal := GpsToDecimal(val)
	if decimal == 0 || GpsRefRegexp.MatchString(val) {
		return decimal
	}
	if isNegativeRef(refChain.firstNonEmpty(root)) {
		return -decimal
	}
	return decimal
}

// Altitude returns the GPS altitude in meters.
// Priority: exif:GPSAltitude (parsed as rational, e.g. "3450/100" → 34.5).
// Composition: exif:GPSAltitudeRef = "1" inverts the sign.
func (doc *XmpDocument) Altitude() float64 {
	val := xmpGPSAltitudeChain.firstNonEmpty(doc.doc)
	if val == "" {
		return 0
	}
	alt, ok := parseRational(val)
	if !ok {
		return 0
	}
	if xmpGPSAltitudeRefChain.firstNonEmpty(doc.doc) == "1" {
		return -alt
	}
	return alt
}

// TakenGps returns the GPS timestamp as a UTC time.Time.
// Priority: exif:GPSTimeStamp (canonical combined ISO 8601 datetime).
// Composition: legacy writers split date/time across exif:GPSDateStamp
// and exif:GPSTimeStamp; the two are joined when the canonical fails.
func (doc *XmpDocument) TakenGps() time.Time {
	ts := xmpGPSTimeStampChain.firstNonEmpty(doc.doc)
	if ts == "" {
		return time.Time{}
	}
	if t := txt.ParseTime(ts, ""); !t.IsZero() {
		return t
	}
	ds := xmpGPSDateStampChain.firstNonEmpty(doc.doc)
	if ds == "" {
		return time.Time{}
	}
	return txt.ParseTime(ds+" "+ts, "")
}

// rationalAccessor reads a rational-valued chain and returns the
// decimal (0 when absent or unparseable). Shared by Aperture/FNumber/FocalLength.
func rationalAccessor(chain chainXPath, root *xmlquery.Node) float64 {
	val := chain.firstNonEmpty(root)
	if val == "" {
		return 0
	}
	if v, ok := parseRational(val); ok {
		return v
	}
	return 0
}

// isNegativeRef returns true when ref starts with S or W (case-insensitive),
// matching the GPSLatitudeRef / GPSLongitudeRef sign convention.
func isNegativeRef(ref string) bool {
	if ref == "" {
		return false
	}
	r := ref[0]
	return r == 'S' || r == 's' || r == 'W' || r == 'w'
}

// parseSubSec converts an EXIF SubSecTime string ("899614") to
// nanoseconds (899,614,000). Inputs longer than 9 digits or
// non-numeric values yield 0.
func parseSubSec(s string) int {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 9 {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	for i := len(s); i < 9; i++ {
		n *= 10
	}
	return n
}

// formatExposure renders a duration as "1/N" for sub-second values
// and as a decimal for one-second-or-longer durations.
func formatExposure(secs float64) string {
	if secs <= 0 {
		return ""
	}
	if secs >= 1 {
		return strconv.FormatFloat(secs, 'f', -1, 64)
	}
	return fmt.Sprintf("1/%d", int(math.Round(1/secs)))
}

// apexToSeconds converts an APEX shutter-speed value (Tv) to seconds
// via t = 2^(-Tv). Input is clamped to [-30, 30] to avoid ±Inf on
// pathological values; the range covers any realistic camera output.
func apexToSeconds(apex float64) float64 {
	if apex < -30 {
		apex = -30
	} else if apex > 30 {
		apex = 30
	}
	return math.Pow(2, -apex)
}

// parseRational parses an XMP rational ("N/D") or plain float.
// The bool distinguishes "parsed as zero" from "could not parse".
func parseRational(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if i := strings.IndexByte(s, '/'); i >= 0 {
		num, err1 := strconv.ParseFloat(strings.TrimSpace(s[:i]), 64)
		den, err2 := strconv.ParseFloat(strings.TrimSpace(s[i+1:]), 64)
		if err1 != nil || err2 != nil || den == 0 {
			return 0, false
		}
		return num / den, true
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return 0, false
}
