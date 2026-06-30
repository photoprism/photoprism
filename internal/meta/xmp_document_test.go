package meta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/antchfx/xmlquery"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

// parseXMPString parses an in-memory XMP string and returns the root
// xmlquery node. Test helper that keeps focused unit tests free of
// temp-file plumbing.
func parseXMPString(t *testing.T, xmp string) *xmlquery.Node {
	t.Helper()
	n, err := xmlquery.Parse(strings.NewReader(xmp))
	if err != nil {
		t.Fatalf("parseXMPString: %v", err)
	}
	return n
}

// loadXmp loads an XMP fixture from path and returns the populated
// XmpDocument; t.Fatal on error. Used by accessor tests that exercise
// the full Load → parse → DOM pipeline.
func loadXmp(t *testing.T, path string) *XmpDocument {
	t.Helper()
	var doc XmpDocument
	if err := doc.Load(path); err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	return &doc
}

// loadXmpString writes body to a temp file in t.TempDir(), loads it
// through Load, and returns the populated XmpDocument. Avoids
// inline-fixture boilerplate in tests that synthesize XMP on the fly.
func loadXmpString(t *testing.T, body string) *XmpDocument {
	t.Helper()
	tmp := filepath.Join(t.TempDir(), "synthetic.xmp")
	if err := os.WriteFile(tmp, []byte(body), fs.ModeFile); err != nil {
		t.Fatalf("write temp xmp: %v", err)
	}
	return loadXmp(t, tmp)
}

func TestMaxDepth(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, 0, maxDepth(nil, 0))
	})
	t.Run("FlatRoot", func(t *testing.T) {
		n := parseXMPString(t, `<?xml version="1.0"?><a/>`)
		// xmlquery wraps the root in a document node; element <a/> is
		// at depth 1 from the document.
		assert.Equal(t, 1, maxDepth(n, 0))
	})
	t.Run("NestedFiveDeep", func(t *testing.T) {
		n := parseXMPString(t, `<?xml version="1.0"?><a><b><c><d><e/></d></c></b></a>`)
		assert.Equal(t, 5, maxDepth(n, 0))
	})
	t.Run("MixedSiblingsTakesDeepestBranch", func(t *testing.T) {
		// Left branch: depth 2; right branch: depth 4. Reader must
		// return the deeper of the two, not the first encountered.
		n := parseXMPString(t, `<?xml version="1.0"?><a><b/><c><d><e><f/></e></d></c></a>`)
		assert.Equal(t, 5, maxDepth(n, 0))
	})
	t.Run("IgnoresTextAndCommentNodes", func(t *testing.T) {
		// Comments and text must not inflate the count.
		n := parseXMPString(t, `<?xml version="1.0"?><a>some text<!-- comment --><b/></a>`)
		assert.Equal(t, 2, maxDepth(n, 0))
	})
}

func TestMustCompile(t *testing.T) {
	t.Run("ValidExpression", func(t *testing.T) {
		expr := mustCompile("//dc:title")
		assert.NotNil(t, expr)
	})
	t.Run("PanicsOnGarbage", func(t *testing.T) {
		assert.Panics(t, func() {
			mustCompile("not [[ valid")
		})
	})
	t.Run("PanicsOnUnknownPrefix", func(t *testing.T) {
		assert.Panics(t, func() {
			mustCompile("//bogus:Element")
		})
	})
}

func TestChainXPath_FirstNonEmpty(t *testing.T) {
	doc := parseXMPString(t, `
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/">
   <dc:title>
    <rdf:Alt>
     <rdf:li xml:lang="x-default">From dc:title</rdf:li>
    </rdf:Alt>
   </dc:title>
   <photoshop:Headline>From Headline</photoshop:Headline>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

	t.Run("FirstLinkWins", func(t *testing.T) {
		// First link in chain matches and returns its value.
		c := chainXPath{
			mustCompile("//dc:title/rdf:Alt/rdf:li[@xml:lang='x-default']"),
			mustCompile("//photoshop:Headline"),
		}
		assert.Equal(t, "From dc:title", c.firstNonEmpty(doc))
	})
	t.Run("FallsThroughOnNoMatch", func(t *testing.T) {
		// First link does not match, second link wins.
		c := chainXPath{
			mustCompile("//dc:nonexistent"),
			mustCompile("//photoshop:Headline"),
		}
		assert.Equal(t, "From Headline", c.firstNonEmpty(doc))
	})
	t.Run("ReturnsEmptyWhenNoLinkMatches", func(t *testing.T) {
		c := chainXPath{
			mustCompile("//dc:nonexistent"),
			mustCompile("//tiff:Make"),
		}
		assert.Equal(t, "", c.firstNonEmpty(doc))
	})
	t.Run("HandlesNilRoot", func(t *testing.T) {
		c := chainXPath{mustCompile("//dc:title")}
		assert.Equal(t, "", c.firstNonEmpty(nil))
	})
	t.Run("HandlesEmptyChain", func(t *testing.T) {
		c := chainXPath{}
		assert.Equal(t, "", c.firstNonEmpty(doc))
	})
}

func TestQueryAll(t *testing.T) {
	doc := parseXMPString(t, `
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/">
   <dc:subject>
    <rdf:Bag>
     <rdf:li>One</rdf:li>
     <rdf:li>Two</rdf:li>
     <rdf:li>  </rdf:li>
     <rdf:li>Three</rdf:li>
    </rdf:Bag>
   </dc:subject>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)

	t.Run("ReturnsAllMatches", func(t *testing.T) {
		got := queryAll(doc, mustCompile("//dc:subject/rdf:Bag/rdf:li"))
		assert.Equal(t, []string{"One", "Two", "Three"}, got)
	})
	t.Run("DropsEmptyTrimmedMatches", func(t *testing.T) {
		// The third <rdf:li> contains only whitespace and is dropped.
		got := queryAll(doc, mustCompile("//dc:subject/rdf:Bag/rdf:li"))
		assert.NotContains(t, got, "")
		assert.Len(t, got, 3)
	})
	t.Run("ReturnsNilForNilRoot", func(t *testing.T) {
		assert.Nil(t, queryAll(nil, mustCompile("//dc:subject")))
	})
	t.Run("ReturnsNilForNilExpression", func(t *testing.T) {
		assert.Nil(t, queryAll(doc, nil))
	})
	t.Run("EmptyMatchSetReturnsEmptySlice", func(t *testing.T) {
		got := queryAll(doc, mustCompile("//dc:nonexistent"))
		assert.Empty(t, got)
	})
}

func TestXmpDocument_Load(t *testing.T) {
	t.Run("MissingFile", func(t *testing.T) {
		var doc XmpDocument
		err := doc.Load("testdata/does-not-exist.xmp")
		assert.Error(t, err)
	})
	t.Run("MalformedXML", func(t *testing.T) {
		// Write a malformed XMP to a temp file and assert Load surfaces
		// the parse error rather than silently succeeding.
		tmp := filepath.Join(t.TempDir(), "broken.xmp")
		if err := writeFile(t, tmp, "<?xml version=\"1.0\"?><a><b></a>"); err != nil {
			t.Fatal(err)
		}
		var doc XmpDocument
		err := doc.Load(tmp)
		assert.Error(t, err)
	})
	t.Run("WellFormedSidecar", func(t *testing.T) {
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.NotNil(t, doc.doc)
	})
}

// writeFile writes content to path with mode fs.ModeFile. Returns the
// underlying error so tests can assert on it.
func writeFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), fs.ModeFile)
}

func TestXmpDocument_TitleAltLanguageFallback(t *testing.T) {
	t.Run("PrefersXDefault", func(t *testing.T) {
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.Equal(t, "Night Shift / Berlin / 2020", doc.Title())
	})
	t.Run("FallsBackToFirstLiWhenNoXDefault", func(t *testing.T) {
		// alt-edge-cases.xmp has dc:title with only de + en (no x-default).
		// The chain must take the first li (de = "Sonnenuntergang").
		doc := loadXmp(t, "testdata/xmp/synthetic/alt-edge-cases.xmp")
		assert.Equal(t, "Sonnenuntergang", doc.Title())
	})
	t.Run("FallsBackToBareText", func(t *testing.T) {
		// apple-test-2.xmp has <dc:title>Botanischer Garten</dc:title>
		// without rdf:Alt; the chain falls through to the bare-text branch.
		doc := loadXmp(t, "testdata/apple-test-2.xmp")
		assert.Equal(t, "Botanischer Garten", doc.Title())
	})
	t.Run("FallsBackToPhotoshopHeadline", func(t *testing.T) {
		// Synthetic case: no dc:title, only photoshop:Headline. The
		// chain's last link must trigger.
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/">
   <photoshop:Headline>Headline-only Title</photoshop:Headline>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.Equal(t, "Headline-only Title", doc.Title())
	})
}

func TestXmpDocument_SubjectBagAndSeq(t *testing.T) {
	t.Run("BagForm", func(t *testing.T) {
		// digikam fixture writes dc:subject as <rdf:Bag>, which the old
		// reader silently dropped. Subject() must now read it.
		doc := loadXmp(t, "testdata/xmp/digikam/aurora.jpg.xmp")
		got := doc.Subject()
		assert.Contains(t, got, "Nature")
		assert.Contains(t, got, "Iceland")
		assert.Contains(t, got, "Aurora")
	})
	t.Run("SeqForm", func(t *testing.T) {
		// Synthetic Seq fixture confirms the legacy form still works.
		doc := loadXmp(t, "testdata/xmp/synthetic/subject-seq.xmp")
		assert.Equal(t, "Sequenced, Keywords, Should, Also, Parse", doc.Subject())
	})
	t.Run("EmptyForMissingSubject", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.Subject())
	})
}

func TestXmpDocument_Subject(t *testing.T) {
	t.Run("ReadsDcSubject", func(t *testing.T) {
		// dc:subject is the primary Subject source and present in most tagged
		// files. It feeds Details.Subject, not the keyword list.
		doc := loadXmp(t, "testdata/xmp/darktable/aurora.jpg.xmp")
		assert.Contains(t, doc.Subject(), "Aurora")
	})
	t.Run("PersonInImageFallback", func(t *testing.T) {
		// No dc:subject → Subject falls back to Iptc4xmpExt:PersonInImage.
		doc := loadXmpString(t, `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:Iptc4xmpExt="http://iptc.org/std/Iptc4xmpExt/2008-02-29/">
   <Iptc4xmpExt:PersonInImage>
    <rdf:Bag>
     <rdf:li>Alice</rdf:li>
     <rdf:li>Bob</rdf:li>
    </rdf:Bag>
   </Iptc4xmpExt:PersonInImage>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.Equal(t, "Alice, Bob", doc.Subject())
	})
	t.Run("HierarchicalSubjectFallback", func(t *testing.T) {
		// No dc:subject, no PersonInImage → lr:hierarchicalSubject.
		doc := loadXmpString(t, `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="" xmlns:lr="http://ns.adobe.com/lightroom/1.0/">
   <lr:hierarchicalSubject>
    <rdf:Bag>
     <rdf:li>Nature|Animals</rdf:li>
     <rdf:li>Places</rdf:li>
    </rdf:Bag>
   </lr:hierarchicalSubject>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.Equal(t, "Nature|Animals, Places", doc.Subject())
	})
	t.Run("EmptyWhenNoSource", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.Subject())
	})
}

func TestXmpDocument_FavoriteAttribute(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.True(t, doc.Favorite())
	})
	t.Run("FalseWhenAttributeAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.False(t, doc.Favorite())
	})
	t.Run("FalseWhenDocumentNotLoaded", func(t *testing.T) {
		var doc XmpDocument
		assert.False(t, doc.Favorite())
	})
}

func TestXmpDocument_MultiRdfDescription(t *testing.T) {
	// multi-rdf-description.xmp splits properties across four sibling
	// <rdf:Description> blocks. The XPath reader must walk them all.
	doc := loadXmp(t, "testdata/xmp/synthetic/multi-rdf-description.xmp")
	assert.Equal(t, "Multi-Description Fixture", doc.Title())
	assert.Equal(t, "PhotoPrism", doc.Artist())
	assert.Equal(t, "SyntheticCam", doc.CameraMake())
	assert.Equal(t, "SC-1 Mark II", doc.CameraModel())
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", doc.DocumentID())
	assert.Equal(t, "xmp.iid:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", doc.InstanceID())
}

func TestXmpDocument_License(t *testing.T) {
	t.Run("DigikamFixture", func(t *testing.T) {
		// digiKam writes xmpRights:UsageTerms as lang-alt with x-default.
		doc := loadXmp(t, "testdata/xmp/digikam/aurora.jpg.xmp")
		assert.Equal(t, "CC-BY-SA 4.0", doc.License())
	})
	t.Run("AdobeBridgeFixture", func(t *testing.T) {
		doc := loadXmp(t, "testdata/xmp/adobe/bridge-2.xmp")
		assert.NotEmpty(t, doc.License())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.Equal(t, "", doc.License())
	})
}

func TestXmpDocument_Software(t *testing.T) {
	t.Run("SyntheticFixture", func(t *testing.T) {
		// software-only.xmp carries only xmp:CreatorTool.
		doc := loadXmp(t, "testdata/xmp/synthetic/software-only.xmp")
		assert.Equal(t, "PhotoPrism Synthetic Fixture 1.0.0", doc.Software())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.Software())
	})
}

func TestXmpDocument_DocumentID(t *testing.T) {
	t.Run("DigikamPrefersOriginalDocumentID", func(t *testing.T) {
		// digiKam writes all three IDs; the chain must pick
		// xmpMM:OriginalDocumentID — the asset-stable identifier.
		doc := loadXmp(t, "testdata/xmp/digikam/aurora.jpg.xmp")
		assert.Equal(t, "254738CA43CD69C01101874D65B006B4", doc.DocumentID())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.DocumentID())
	})
}

func TestXmpDocument_InstanceID(t *testing.T) {
	t.Run("DigikamFixture", func(t *testing.T) {
		doc := loadXmp(t, "testdata/xmp/digikam/aurora.jpg.xmp")
		assert.Equal(t, "xmp.iid:de0fb44d-bdc8-4cc3-aa50-9aefe4992b34", doc.InstanceID())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.InstanceID())
	})
}

func TestXmpDocument_GPSCoordinates(t *testing.T) {
	t.Run("AdobeTwoComponentForm", func(t *testing.T) {
		// bridge-2 carries lat/lng/alt in 2-component form (Adobe XMP).
		doc := loadXmp(t, "testdata/xmp/adobe/bridge-2.xmp")
		assert.InEpsilon(t, 52.44308166666666, doc.Lat(), 1e-6)
		assert.InEpsilon(t, 13.576613333333333, doc.Lng(), 1e-6)
		// 1500000/100 = 15000 m above sea level.
		assert.InEpsilon(t, 15000.0, doc.Altitude(), 1e-6)
	})
	t.Run("SouthernHemisphereInValue", func(t *testing.T) {
		// bridge.xmp Lat is "27,20.4263S" — sign comes from cardinal in
		// the value string, not GPSLatitudeRef.
		doc := loadXmp(t, "testdata/xmp/adobe/bridge.xmp")
		assert.Less(t, doc.Lat(), 0.0)
		assert.InEpsilon(t, -27.340438333333333, doc.Lat(), 1e-6)
	})
	t.Run("DecimalWithSeparateRef", func(t *testing.T) {
		// Synthetic case: pure-decimal value with sign supplied by the
		// GPSLatitudeRef element. The accessor must apply the sign.
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:GPSLatitude>47.6754</exif:GPSLatitude>
   <exif:GPSLatitudeRef>S</exif:GPSLatitudeRef>
   <exif:GPSLongitude>122.3127</exif:GPSLongitude>
   <exif:GPSLongitudeRef>W</exif:GPSLongitudeRef>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.Equal(t, -47.6754, doc.Lat())
		assert.Equal(t, -122.3127, doc.Lng())
	})
	t.Run("AltitudeBelowSeaLevel", func(t *testing.T) {
		// GPSAltitudeRef="1" inverts the sign.
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:GPSAltitude>10000/100</exif:GPSAltitude>
   <exif:GPSAltitudeRef>1</exif:GPSAltitudeRef>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.Equal(t, -100.0, doc.Altitude())
	})
	t.Run("ZeroWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, 0.0, doc.Lat())
		assert.Equal(t, 0.0, doc.Lng())
		assert.Equal(t, 0.0, doc.Altitude())
	})
}

func TestXmpDocument_TakenGps(t *testing.T) {
	t.Run("CombinedISO8601Form", func(t *testing.T) {
		// Canonical XMP: GPSTimeStamp is the full ISO 8601 datetime.
		doc := loadXmp(t, "testdata/xmp/synthetic/gps-time-combined.xmp")
		got := doc.TakenGps()
		assert.False(t, got.IsZero())
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, 5, int(got.Month()))
		assert.Equal(t, 6, got.Day())
	})
	t.Run("LegacySplitForm", func(t *testing.T) {
		// Older writers split date and time across two elements.
		doc := loadXmp(t, "testdata/xmp/synthetic/gps-time-split.xmp")
		got := doc.TakenGps()
		assert.False(t, got.IsZero(), "split-form GPSTimeStamp should still parse")
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, 5, int(got.Month()))
		assert.Equal(t, 6, got.Day())
	})
	t.Run("ZeroWhenAbsent", func(t *testing.T) {
		// canon_eos_6d carries timestamps but no GPS time.
		doc := loadXmp(t, "testdata/canon_eos_6d.xmp")
		assert.True(t, doc.TakenGps().IsZero())
	})
}

func TestXmpDocument_CopyrightWebStatementFallback(t *testing.T) {
	// When dc:rights is absent but xmpRights:WebStatement is present,
	// the chain falls through to the URL.
	doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:xmpRights="http://ns.adobe.com/xap/1.0/rights/">
   <xmpRights:WebStatement>https://example.org/license</xmpRights:WebStatement>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
	assert.Equal(t, "https://example.org/license", doc.Copyright())
}

func TestParseRational(t *testing.T) {
	t.Run("RationalWithDenominator", func(t *testing.T) {
		v, ok := parseRational("3450/100")
		assert.True(t, ok)
		assert.Equal(t, 34.5, v)
	})
	t.Run("PureFloat", func(t *testing.T) {
		v, ok := parseRational("47.6754")
		assert.True(t, ok)
		assert.Equal(t, 47.6754, v)
	})
	t.Run("ZeroNumerator", func(t *testing.T) {
		v, ok := parseRational("0/100")
		assert.True(t, ok)
		assert.Equal(t, 0.0, v)
	})
	t.Run("DivisionByZero", func(t *testing.T) {
		_, ok := parseRational("1/0")
		assert.False(t, ok)
	})
	t.Run("Empty", func(t *testing.T) {
		_, ok := parseRational("")
		assert.False(t, ok)
	})
	t.Run("Garbage", func(t *testing.T) {
		_, ok := parseRational("abc/def")
		assert.False(t, ok)
	})
	t.Run("Whitespace", func(t *testing.T) {
		v, ok := parseRational("  100/4  ")
		assert.True(t, ok)
		assert.Equal(t, 25.0, v)
	})
}

func TestIsNegativeRef(t *testing.T) {
	assert.True(t, isNegativeRef("S"))
	assert.True(t, isNegativeRef("s"))
	assert.True(t, isNegativeRef("W"))
	assert.True(t, isNegativeRef("w"))
	assert.True(t, isNegativeRef("South"), "prefix-based check accepts 'South'")
	assert.False(t, isNegativeRef("N"))
	assert.False(t, isNegativeRef("E"))
	assert.False(t, isNegativeRef(""))
	assert.False(t, isNegativeRef("  "))
}

func TestParseSubSec(t *testing.T) {
	t.Run("SixDigits", func(t *testing.T) {
		// "899614" represents 0.899614 s = 899,614,000 ns.
		assert.Equal(t, 899614000, parseSubSec("899614"))
	})
	t.Run("NineDigits", func(t *testing.T) {
		// Nanosecond-precision input is returned unchanged.
		assert.Equal(t, 729626112, parseSubSec("729626112"))
	})
	t.Run("ThreeDigits", func(t *testing.T) {
		// "123" represents 0.123 s = 123,000,000 ns.
		assert.Equal(t, 123000000, parseSubSec("123"))
	})
	t.Run("Whitespace", func(t *testing.T) {
		assert.Equal(t, 500000000, parseSubSec("  5  "))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, 0, parseSubSec(""))
	})
	t.Run("TooLongRejected", func(t *testing.T) {
		// >9 digits would overflow nanosecond precision.
		assert.Equal(t, 0, parseSubSec("1234567890"))
	})
	t.Run("NonNumericRejected", func(t *testing.T) {
		assert.Equal(t, 0, parseSubSec("abc"))
	})
}

func TestFormatExposure(t *testing.T) {
	t.Run("FractionalSeconds", func(t *testing.T) {
		assert.Equal(t, "1/250", formatExposure(0.004))
		assert.Equal(t, "1/50", formatExposure(0.02))
		assert.Equal(t, "1/2", formatExposure(0.5))
	})
	t.Run("OneSecondAndAbove", func(t *testing.T) {
		assert.Equal(t, "1", formatExposure(1.0))
		assert.Equal(t, "2.5", formatExposure(2.5))
		assert.Equal(t, "30", formatExposure(30))
	})
	t.Run("ZeroOrNegative", func(t *testing.T) {
		assert.Equal(t, "", formatExposure(0))
		assert.Equal(t, "", formatExposure(-1))
	})
}

func TestApexToSeconds(t *testing.T) {
	// EXIF spec: Tv = 0 → 1 s, Tv = 5 → 1/32 s.
	assert.InDelta(t, 1.0, apexToSeconds(0), 1e-9)
	assert.InDelta(t, 0.5, apexToSeconds(1), 1e-9)
	assert.InDelta(t, 1.0/32.0, apexToSeconds(5), 1e-9)
	// Negative APEX = exposures longer than 1 second.
	assert.InDelta(t, 2.0, apexToSeconds(-1), 1e-9)
}

func TestXmpDocument_CameraSerial(t *testing.T) {
	t.Run("PrefersExifEXSerialNumber", func(t *testing.T) {
		// Synthetic exifEX fixture writes both exifEX:SerialNumber and
		// (no aux:SerialNumber). Modern fallback wins.
		doc := loadXmp(t, "testdata/xmp/synthetic/exifex-camera-lens.xmp")
		assert.Equal(t, "SC1-BODY-123456", doc.CameraSerial())
	})
	t.Run("FallsBackToAuxSerialNumber", func(t *testing.T) {
		// canon_eos_6d has aux:SerialNumber but no exifEX:SerialNumber.
		doc := loadXmp(t, "testdata/xmp/synthetic/aux-only.xmp")
		assert.Equal(t, "BODY-SN-123456", doc.CameraSerial())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.CameraSerial())
	})
}

func TestXmpDocument_LensMake(t *testing.T) {
	t.Run("ExifEXFixture", func(t *testing.T) {
		doc := loadXmp(t, "testdata/xmp/synthetic/exifex-camera-lens.xmp")
		assert.Equal(t, "SyntheticLens Co.", doc.LensMake())
	})
	t.Run("IphoneSeven", func(t *testing.T) {
		// iphone_7 fixture writes <exifEX:LensMake>Apple</exifEX:LensMake>.
		doc := loadXmp(t, "testdata/iphone_7.xmp")
		assert.Equal(t, "Apple", doc.LensMake())
	})
}

func TestXmpDocument_CameraOwner(t *testing.T) {
	doc := loadXmp(t, "testdata/xmp/synthetic/aux-only.xmp")
	assert.Equal(t, "Synthetic Photographer", doc.CameraOwner())
}

func TestXmpDocument_Projection(t *testing.T) {
	t.Run("Equirectangular", func(t *testing.T) {
		doc := loadXmp(t, "testdata/xmp/synthetic/gpano-360.xmp")
		assert.Equal(t, "equirectangular", doc.Projection())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.Equal(t, "", doc.Projection())
	})
}

func TestXmpDocument_ColorProfile(t *testing.T) {
	doc := loadXmp(t, "testdata/photoshop.xmp")
	assert.Equal(t, "sRGB IEC61966-2.1", doc.ColorProfile())
}

func TestXmpDocument_Aperture(t *testing.T) {
	t.Run("Photoshop", func(t *testing.T) {
		// photoshop.xmp has <exif:ApertureValue>1695994/1000000</exif:ApertureValue>.
		doc := loadXmp(t, "testdata/photoshop.xmp")
		// 1695994/1000000 ≈ 1.696
		assert.InDelta(t, 1.696, doc.Aperture(), 0.01)
	})
	t.Run("ZeroWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, float32(0), doc.Aperture())
	})
}

func TestXmpDocument_FNumber(t *testing.T) {
	// photoshop.xmp has <exif:FNumber>180/100</exif:FNumber> = 1.8.
	doc := loadXmp(t, "testdata/photoshop.xmp")
	assert.InDelta(t, float32(1.8), doc.FNumber(), 0.01)
}

func TestXmpDocument_FocalLength(t *testing.T) {
	// photoshop.xmp has <exif:FocalLength>5580/1000</exif:FocalLength> ≈ 5.58 mm → 6.
	doc := loadXmp(t, "testdata/photoshop.xmp")
	assert.Equal(t, 6, doc.FocalLength())
}

func TestXmpDocument_Iso(t *testing.T) {
	t.Run("FromISOSpeedRatingsSeq", func(t *testing.T) {
		// photoshop.xmp has <exif:ISOSpeedRatings><rdf:Seq><rdf:li>200</rdf:li>...
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.Equal(t, 200, doc.Iso())
	})
	t.Run("FromExifEXPhotographicSensitivity", func(t *testing.T) {
		// Synthetic exifEX fixture exposes the modern element.
		doc := loadXmp(t, "testdata/xmp/synthetic/exifex-camera-lens.xmp")
		assert.Equal(t, 800, doc.Iso())
	})
}

func TestXmpDocument_Exposure(t *testing.T) {
	t.Run("ExposureTimeRational", func(t *testing.T) {
		// photoshop.xmp has 20000000/1000000000 = 1/50 s.
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.Equal(t, "1/50", doc.Exposure())
	})
	t.Run("ShutterSpeedFallback", func(t *testing.T) {
		// Synthetic case: only ShutterSpeedValue present. APEX 5 → 1/32 s.
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:ShutterSpeedValue>5/1</exif:ShutterSpeedValue>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.Equal(t, "1/32", doc.Exposure())
	})
}

func TestXmpDocument_Flash(t *testing.T) {
	t.Run("FalseFromStruct", func(t *testing.T) {
		// photoshop.xmp has <exif:Flash><exif:Fired>False</exif:Fired>...
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.False(t, doc.Flash())
	})
	t.Run("TrueFromStruct", func(t *testing.T) {
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:Flash rdf:parseType="Resource">
    <exif:Fired>True</exif:Fired>
    <exif:Mode>1</exif:Mode>
   </exif:Flash>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		assert.True(t, doc.Flash())
	})
	t.Run("FalseWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.False(t, doc.Flash())
	})
}

func TestXmpDocument_Notes(t *testing.T) {
	t.Run("LangAlt", func(t *testing.T) {
		doc := loadXmp(t, "testdata/xmp/synthetic/notes-usercomment.xmp")
		assert.Contains(t, doc.Notes(), "Notes accessor")
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.Notes())
	})
}

func TestXmpDocument_TakenNs(t *testing.T) {
	t.Run("FromSubSecTimeOriginal", func(t *testing.T) {
		// photoshop.xmp has <exif:SubSecTimeOriginal>899614</exif:SubSecTimeOriginal>.
		doc := loadXmp(t, "testdata/photoshop.xmp")
		assert.Equal(t, 899614000, doc.TakenNs())
	})
	t.Run("ZeroWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, 0, doc.TakenNs())
	})
}

func TestXmpDocument_TakenAtChainAndSubSecJoin(t *testing.T) {
	t.Run("PreservesFractionalFromDateString", func(t *testing.T) {
		// photoshop.xmp's photoshop:DateCreated already has fractional
		// seconds; the SubSec join must NOT override.
		doc := loadXmp(t, "testdata/photoshop.xmp")
		got := doc.TakenAt("")
		assert.Equal(t, 729626112, got.Nanosecond())
	})
	t.Run("JoinsSubSecWhenDateLacksFraction", func(t *testing.T) {
		// Synthetic case: DateCreated without fractional + SubSec.
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/"
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <photoshop:DateCreated>2020-01-01T17:28:23</photoshop:DateCreated>
   <exif:SubSecTimeOriginal>123456</exif:SubSecTimeOriginal>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		got := doc.TakenAt("")
		assert.Equal(t, 123456000, got.Nanosecond())
	})
	t.Run("FallsBackToDateTimeOriginal", func(t *testing.T) {
		// No photoshop:DateCreated; the chain must move to
		// exif:DateTimeOriginal.
		doc := loadXmpString(t, `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:DateTimeOriginal>2019-08-15T10:00:00</exif:DateTimeOriginal>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`)
		got := doc.TakenAt("")
		assert.Equal(t, 2019, got.Year())
		assert.Equal(t, 8, int(got.Month()))
	})
}

func TestXmpDocument_CreatedAt(t *testing.T) {
	t.Run("FromXmpCreateDate", func(t *testing.T) {
		// photoshop.xmp has <xmp:CreateDate>2020-01-01T17:28:23</xmp:CreateDate>.
		doc := loadXmp(t, "testdata/photoshop.xmp")
		got := doc.CreatedAt("")
		assert.Equal(t, 2020, got.Year())
	})
	t.Run("FallsBackToXmpDMCreationDate", func(t *testing.T) {
		// xmpDM:CreationDate is only used for video/audio sidecars.
		doc := loadXmp(t, "testdata/xmp/synthetic/xmpdm-creationdate.xmp")
		got := doc.CreatedAt("")
		assert.Equal(t, 2026, got.Year())
		assert.Equal(t, 5, int(got.Month()))
		assert.Equal(t, 6, got.Day())
	})
}

func TestXmpDocument_TimeOffset(t *testing.T) {
	t.Run("OffsetTimeOriginal", func(t *testing.T) {
		doc := loadXmp(t, "testdata/xmp/synthetic/time-offsets-subsec.xmp")
		assert.Equal(t, "+02:00", doc.TimeOffset())
	})
	t.Run("EmptyWhenAbsent", func(t *testing.T) {
		doc := loadXmp(t, "testdata/fstop-favorite.xmp")
		assert.Equal(t, "", doc.TimeOffset())
	})
}

// TestXmpDocument_DarktableFixture covers the Darktable sidecar end to
// end — descriptive metadata, Bag-form keywords, and the 2-component
// Adobe GPS form that motivates the GpsToDecimal extension. The
// fixture also carries `<darktable:history>`, `lr:hierarchicalSubject`,
// and `xmp:Rating`, all of which are out of scope for the current
// reader; the test asserts they are silently ignored rather than
// surfacing as garbage on any in-scope field.
func TestXmpDocument_DarktableFixture(t *testing.T) {
	doc := loadXmp(t, "testdata/xmp/darktable/aurora.jpg.xmp")

	t.Run("DescriptiveMetadata", func(t *testing.T) {
		assert.Equal(t, "XMP Test - Aurora", doc.Title())
		assert.Equal(t, "Test fixture for darktable  XMP sidecar", doc.Description())
		assert.Equal(t, "PhotoPrism", doc.Artist())
		assert.Equal(t, "CC-BY-SA 4.0", doc.Copyright())
	})
	t.Run("BagFormSubject", func(t *testing.T) {
		// Darktable writes <dc:subject><rdf:Bag>. The old reader
		// dropped this entirely (it only handled <rdf:Seq>); the new
		// reader must produce a non-empty Subject list.
		got := doc.Subject()
		assert.Contains(t, got, "Aurora")
		assert.Contains(t, got, "Iceland")
		assert.Contains(t, got, "Nature")
	})
	t.Run("TwoComponentGPSForm", func(t *testing.T) {
		// The fixture carries lat="87,21.291962N" / lng="179,59.546814W" —
		// the canonical regression case for the GpsToDecimal extension.
		// 87° 21.291962'N → 87 + 21.291962/60 ≈ 87.3549.
		assert.InDelta(t, 87.3549, doc.Lat(), 0.001)
		// W cardinal inverts the sign for longitude.
		assert.Less(t, doc.Lng(), 0.0)
		assert.InDelta(t, -179.9924, doc.Lng(), 0.001)
	})
	t.Run("OutOfScopeTagsIgnored", func(t *testing.T) {
		// xmp:Rating is gap-analysis section 2 (rating/triage feature
		// epic). The reader has no accessor for it, so no assertion is
		// possible directly — we instead confirm that none of the
		// in-scope accessors leak the rating value into their output.
		assert.NotContains(t, doc.Title(), "1")
		assert.NotContains(t, doc.Description(), "1")
		// darktable:history blocks must not pollute Software either.
		assert.Equal(t, "", doc.Software())
	})
}
