package meta

import (
	"os"
	"testing"

	"github.com/tidwall/gjson"
)

// parseFaces returns the parsed regions as faces plus the partial flag, keeping
// the region-shape assertions below independent of the FaceRegions wrapper.
func parseFaces(j gjson.Result, options FaceOptions) ([]Face, bool) {
	regions := parseExiftoolFaces(j, options)

	return regions.Faces, regions.Partial
}

// exiftoolFaces reads a fixture ExifTool JSON dump and returns the parsed faces.
func exiftoolFaces(t *testing.T, path string) []Face {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test reads a fixture path
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var data Data
	if err := data.Exiftool(b, ""); err != nil {
		t.Fatalf("exiftool %s: %v", path, err)
	}
	return data.Faces
}

func TestData_Exiftool_Faces_Single(t *testing.T) {
	faces := exiftoolFaces(t, "testdata/faces/mwg-single.json")
	if len(faces) != 1 {
		t.Fatalf("want 1 face, got %d (%+v)", len(faces), faces)
	}
	f := faces[0]
	// center (0.5,0.4) size (0.1,0.15) -> top-left (0.45,0.325); Iceland-P3 orientation 1.
	if f.Name != "Alice" || !almost(f.X, 0.45) || !almost(f.Y, 0.325) || !almost(f.W, 0.1) || !almost(f.H, 0.15) {
		t.Errorf("got %+v", f)
	}
}

func TestData_Exiftool_Faces_Multi(t *testing.T) {
	faces := exiftoolFaces(t, "testdata/faces/mwg-multi.json")
	if len(faces) != 2 {
		t.Fatalf("want 2 faces, got %d (%+v)", len(faces), faces)
	}
	byName := map[string]Face{}
	for _, f := range faces {
		byName[f.Name] = f
	}
	if a, ok := byName["Alice"]; !ok || !almost(a.X, 0.45) || !almost(a.Y, 0.325) {
		t.Errorf("Alice wrong: %+v", a)
	}
	if b, ok := byName["Bob"]; !ok || !almost(b.X, 0.14) || !almost(b.Y, 0.21) {
		// center (0.2,0.3) size (0.12,0.18) -> TL (0.14,0.21)
		t.Errorf("Bob wrong: %+v", b)
	}
}

func TestData_Exiftool_Faces_MP(t *testing.T) {
	faces := exiftoolFaces(t, "testdata/faces/mp.json")
	if len(faces) != 1 {
		t.Fatalf("want 1 face, got %d (%+v)", len(faces), faces)
	}
	f := faces[0]
	if f.Name != "Cara" || !almost(f.X, 0.3) || !almost(f.Y, 0.2) || !almost(f.W, 0.1) || !almost(f.H, 0.15) {
		t.Errorf("got %+v", f)
	}
}

// TestParseExiftoolFaces_Compaction verifies defensive handling of ExifTool's
// non-parallel (compacted) arrays: aligned optional members are used, sparse
// ones are dropped and reported partial, and mixed shapes are skipped rather
// than mis-sized. Shapes mirror real ExifTool "-n -m -j" output.
func TestParseExiftoolFaces_Compaction(t *testing.T) {
	t.Run("AllNamedRectangles", func(t *testing.T) {
		j := gjson.Parse(`{
			"RegionName":["Alice","Bob"],
			"RegionType":["Face","Face"],
			"RegionAreaX":[0.2,0.5],
			"RegionAreaY":[0.5,0.5],
			"RegionAreaW":[0.1,0.1],
			"RegionAreaH":[0.1,0.1]
		}`)
		faces, partial := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if partial {
			t.Error("fully aligned set must not be partial")
		}
		if len(faces) != 2 || faces[0].Name != "Alice" || faces[1].Name != "Bob" {
			t.Fatalf("got %+v", faces)
		}
	})
	t.Run("TypeAbsentImported", func(t *testing.T) {
		// mwg-rs:Type is optional, so a region without one is imported as a face.
		j := gjson.Parse(`{
			"RegionName":["Alice"],
			"RegionAreaX":[0.2],
			"RegionAreaY":[0.5],
			"RegionAreaW":[0.1],
			"RegionAreaH":[0.1]
		}`)
		faces, partial := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if partial {
			t.Error("an absent RegionType resolves every region, so the parse is not partial")
		}
		if len(faces) != 1 || faces[0].Name != "Alice" {
			t.Fatalf("type-less regions must be imported, got %+v", faces)
		}
	})
	t.Run("TypePresentNonFaceSkipped", func(t *testing.T) {
		j := gjson.Parse(`{
			"RegionName":["Alice","Bar"],
			"RegionType":["Face","BarCode"],
			"RegionAreaX":[0.2,0.5],
			"RegionAreaY":[0.5,0.5],
			"RegionAreaW":[0.1,0.1],
			"RegionAreaH":[0.1,0.1]
		}`)
		faces, _ := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if len(faces) != 1 || faces[0].Name != "Alice" {
			t.Fatalf("only the Face-typed region must import, got %+v", faces)
		}
	})
	t.Run("UnnamedRegionDropsNamesNotBoxes", func(t *testing.T) {
		// Real shape: the middle region is unnamed, so RegionName compacts to
		// len 2 while the geometry arrays stay len 3.
		j := gjson.Parse(`{
			"RegionName":["Alice","Bob"],
			"RegionType":["Face","Face","Face"],
			"RegionAreaX":[0.2,0.5,0.8],
			"RegionAreaY":[0.5,0.5,0.5],
			"RegionAreaW":[0.1,0.1,0.1],
			"RegionAreaH":[0.1,0.1,0.1]
		}`)
		faces, partial := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if !partial {
			t.Error("compacted names must report partial")
		}
		if len(faces) != 3 {
			t.Fatalf("want 3 boxes, got %d: %+v", len(faces), faces)
		}
		for i, f := range faces {
			if f.Name != "" {
				t.Errorf("region %d must be unnamed rather than mislabeled, got %q", i, f.Name)
			}
		}
		if !almost(faces[0].X, 0.15) || !almost(faces[2].X, 0.75) {
			t.Errorf("boxes must stay aligned: %+v", faces)
		}
	})
	t.Run("MixedCircleAndRectangleSkipped", func(t *testing.T) {
		// Real shape for a rectangle followed by a circle: W/H compact to len 1
		// and D collapses to a scalar, so the set cannot be disambiguated.
		j := gjson.Parse(`{
			"RegionName":["Alice","Carol"],
			"RegionType":["Face","Face"],
			"RegionAreaX":[0.2,0.7],
			"RegionAreaY":[0.5,0.4],
			"RegionAreaW":[0.1],
			"RegionAreaH":[0.15],
			"RegionAreaD":0.2
		}`)
		faces, partial := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if !partial {
			t.Error("mixed shapes must report partial")
		}
		if len(faces) != 0 {
			t.Fatalf("mixed shapes must be skipped, got %+v", faces)
		}
	})
	t.Run("AllCircles", func(t *testing.T) {
		j := gjson.Parse(`{
			"RegionName":["Alice","Bob"],
			"RegionType":["Face","Face"],
			"RegionAreaX":[0.5,0.5],
			"RegionAreaY":[0.5,0.5],
			"RegionAreaD":[0.2,0.2]
		}`)
		faces, partial := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if partial {
			t.Error("uniform circles must not be partial")
		}
		if len(faces) != 2 || !almost(faces[0].W, 0.1) || !almost(faces[0].H, 0.2) {
			t.Fatalf("got %+v", faces)
		}
	})
	t.Run("SingleRegionNameFallbacks", func(t *testing.T) {
		title := gjson.Parse(`{
			"RegionTitle":"Title Person","RegionType":"Face",
			"RegionAreaX":0.5,"RegionAreaY":0.5,"RegionAreaW":0.1,"RegionAreaH":0.1
		}`)
		faces, partial := parseFaces(title, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if partial || len(faces) != 1 || faces[0].Name != "Title Person" {
			t.Errorf("title fallback got %+v (partial %t)", faces, partial)
		}
		seeAlso := gjson.Parse(`{
			"RegionSeeAlso":"Iptc4xmpExt:PersonInImage","PersonInImage":"Referenced Person",
			"RegionType":"Face","RegionAreaX":0.5,"RegionAreaY":0.5,"RegionAreaW":0.1,"RegionAreaH":0.1
		}`)
		faces, partial = parseFaces(seeAlso, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if partial || len(faces) != 1 || faces[0].Name != "Referenced Person" {
			t.Errorf("seeAlso fallback got %+v (partial %t)", faces, partial)
		}
	})
	t.Run("ACDSeeDLYPreferred", func(t *testing.T) {
		j := gjson.Parse(`{
			"ACDSeeRegionName":["DLY","ALG"],
			"ACDSeeRegionType":["Face","Face"],
			"ACDSeeRegionDLYAreaX":[0.3,0.6],
			"ACDSeeRegionDLYAreaY":[0.4,0.6],
			"ACDSeeRegionDLYAreaW":[0.2,0.2],
			"ACDSeeRegionDLYAreaH":[0.3,0.2],
			"ACDSeeRegionALGAreaX":[0.8,0.7],
			"ACDSeeRegionALGAreaY":[0.8,0.6],
			"ACDSeeRegionALGAreaW":[0.1,0.1],
			"ACDSeeRegionALGAreaH":[0.1,0.2]
		}`)
		faces, partial := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if partial {
			t.Error("aligned ACDSee set must not be partial")
		}
		if len(faces) != 2 || faces[0].Name != "DLY" || !almost(faces[0].X, 0.2) || !almost(faces[0].Y, 0.25) {
			t.Fatalf("got %+v", faces)
		}
	})
	t.Run("ACDSeeSparseDLYFallsBackToALG", func(t *testing.T) {
		// Only the first region has a DLY area, so DLY compacts to len 1 and the
		// automatic ALG area is used for every region instead.
		j := gjson.Parse(`{
			"ACDSeeRegionName":["First","Second"],
			"ACDSeeRegionType":["Face","Face"],
			"ACDSeeRegionDLYAreaX":[0.3],
			"ACDSeeRegionDLYAreaY":[0.4],
			"ACDSeeRegionDLYAreaW":[0.2],
			"ACDSeeRegionDLYAreaH":[0.3],
			"ACDSeeRegionALGAreaX":[0.8,0.7],
			"ACDSeeRegionALGAreaY":[0.8,0.6],
			"ACDSeeRegionALGAreaW":[0.1,0.1],
			"ACDSeeRegionALGAreaH":[0.1,0.2]
		}`)
		faces, _ := parseFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if len(faces) != 2 || !almost(faces[0].X, 0.75) || !almost(faces[0].Y, 0.75) {
			t.Fatalf("expected ALG fallback boxes, got %+v", faces)
		}
	})
}

// TestData_Exiftool_Faces_Unnamed verifies the end-to-end pipeline on a genuine
// ExifTool dump where an unnamed middle region compacts RegionName: boxes import
// unnamed and the result is flagged partial.
func TestData_Exiftool_Faces_Unnamed(t *testing.T) {
	b, err := os.ReadFile("testdata/faces/mwg-unnamed.json")
	if err != nil {
		t.Fatal(err)
	}
	var data Data
	if err := data.Exiftool(b, ""); err != nil {
		t.Fatal(err)
	}
	if !data.FacesPartial {
		t.Error("compacted names must set FacesPartial")
	}
	if len(data.Faces) != 3 {
		t.Fatalf("want 3 faces, got %d: %+v", len(data.Faces), data.Faces)
	}
	for _, f := range data.Faces {
		if f.Name != "" {
			t.Errorf("region must import unnamed, got %q", f.Name)
		}
	}
}

// TestJsonFaceName covers the per-region name fallback chain.
func TestJsonFaceName(t *testing.T) {
	j := gjson.Parse(`{"PersonInImage":"Solo Person"}`)
	if got := jsonFaceName(j, "Name", "Title", "Ext", ""); got != "Name" {
		t.Errorf("name precedence: got %q", got)
	}
	if got := jsonFaceName(j, "", "Title", "Ext", ""); got != "Title" {
		t.Errorf("title fallback: got %q", got)
	}
	if got := jsonFaceName(j, "", "", "Ext", ""); got != "Ext" {
		t.Errorf("extension fallback: got %q", got)
	}
	if got := jsonFaceName(j, "", "", "", "Iptc4xmpExt:PersonInImage"); got != "Solo Person" {
		t.Errorf("seeAlso fallback: got %q", got)
	}
	multi := gjson.Parse(`{"PersonInImage":["A","B"]}`)
	if got := jsonFaceName(multi, "", "", "", "Iptc4xmpExt:PersonInImage"); got != "" {
		t.Errorf("ambiguous seeAlso must resolve empty, got %q", got)
	}
	if got := jsonFaceName(j, "", "", "", "other"); got != "" {
		t.Errorf("non-matching reference must resolve empty, got %q", got)
	}
}

// TestJsonArrayFit covers per-region array alignment classification.
func TestJsonArrayFit(t *testing.T) {
	three := gjson.Parse(`[1,2,3]`).Array()
	if a, s := jsonArrayFit(three, 3); !a || s {
		t.Errorf("len==n must be aligned: %t %t", a, s)
	}
	if a, s := jsonArrayFit(nil, 3); a || s {
		t.Errorf("empty must be absent, not sparse: %t %t", a, s)
	}
	if a, s := jsonArrayFit(three, 4); a || !s {
		t.Errorf("len<n must be sparse: %t %t", a, s)
	}
}

// TestParseExiftoolFaces_Declared verifies the authoritative-set flags and the
// ACDSee region count, which must not be derived from the optional type array.
func TestParseExiftoolFaces_Declared(t *testing.T) {
	t.Run("NoRegions", func(t *testing.T) {
		regions := parseExiftoolFaces(gjson.Parse(`{"Make":"NIKON"}`), FaceOptions{Orientation: 1})
		if regions.Declared || regions.Partial || len(regions.Faces) != 0 {
			t.Fatalf("a file without regions must not be declared: %+v", regions)
		}
	})
	t.Run("MwgRegionsDeclared", func(t *testing.T) {
		j := gjson.Parse(`{
			"RegionName":["Alice"],
			"RegionType":["Face"],
			"RegionAreaX":[0.2],
			"RegionAreaY":[0.5],
			"RegionAreaW":[0.1],
			"RegionAreaH":[0.1]
		}`)
		regions := parseExiftoolFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if !regions.Declared || regions.Partial || len(regions.Faces) != 1 {
			t.Fatalf("got %+v", regions)
		}
	})
	t.Run("AcdSeeTypeAbsentImported", func(t *testing.T) {
		// ACDSeeRegionType is optional, so the region count must come from the
		// coordinate arrays; deriving it from the type array reports zero regions.
		j := gjson.Parse(`{
			"ACDSeeRegionName":["Dana"],
			"ACDSeeRegionDLYAreaX":[0.3],
			"ACDSeeRegionDLYAreaY":[0.4],
			"ACDSeeRegionDLYAreaW":[0.2],
			"ACDSeeRegionDLYAreaH":[0.3],
			"ACDSeeRegionAppliedToDimensionsW":4000,
			"ACDSeeRegionAppliedToDimensionsH":2000
		}`)
		regions := parseExiftoolFaces(j, FaceOptions{Orientation: 1})
		if !regions.Declared || regions.Partial {
			t.Fatalf("got %+v", regions)
		}
		if len(regions.Faces) != 1 || regions.Faces[0].Name != "Dana" {
			t.Fatalf("type-less ACDSee regions must be imported, got %+v", regions.Faces)
		}
	})
	t.Run("AcdSeeCompactedMemberPartial", func(t *testing.T) {
		// Two typed regions but only one DLY/ALG coordinate set: ExifTool compacted
		// a sparse member, so nothing may be assigned positionally.
		j := gjson.Parse(`{
			"ACDSeeRegionName":["Dana","Eve"],
			"ACDSeeRegionType":["Face","Face"],
			"ACDSeeRegionDLYAreaX":[0.3],
			"ACDSeeRegionDLYAreaY":[0.4],
			"ACDSeeRegionDLYAreaW":[0.2],
			"ACDSeeRegionDLYAreaH":[0.3],
			"ACDSeeRegionAppliedToDimensionsW":4000,
			"ACDSeeRegionAppliedToDimensionsH":2000
		}`)
		regions := parseExiftoolFaces(j, FaceOptions{Orientation: 1})
		if !regions.Declared || !regions.Partial {
			t.Fatalf("a compacted ACDSee member must be declared and partial: %+v", regions)
		}
		if len(regions.Faces) != 0 {
			t.Fatalf("got %+v", regions.Faces)
		}
	})
	t.Run("UnresolvableRegionPartial", func(t *testing.T) {
		// Pixel units with no applied or source dimensions cannot be normalized.
		j := gjson.Parse(`{
			"RegionName":["Alice"],
			"RegionType":["Face"],
			"RegionAreaX":[2000],
			"RegionAreaY":[1500],
			"RegionAreaW":[400],
			"RegionAreaH":[600],
			"RegionAreaUnit":["pixel"]
		}`)
		regions := parseExiftoolFaces(j, FaceOptions{Orientation: 1})
		if !regions.Declared || !regions.Partial || len(regions.Faces) != 0 {
			t.Fatalf("got %+v", regions)
		}
	})
}
