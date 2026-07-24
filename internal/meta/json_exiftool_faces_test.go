package meta

import (
	"os"
	"testing"

	"github.com/tidwall/gjson"
)

// exiftoolFaces reads a fixture ExifTool JSON dump and returns the parsed faces.
func exiftoolFaces(t *testing.T, path string) []Face {
	t.Helper()
	b, err := os.ReadFile(path)
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

// TestParseExiftoolFaces_EdgeCases verifies mirrored, circular, rotated, and ACDSee regions.
func TestParseExiftoolFaces_EdgeCases(t *testing.T) {
	t.Run("MirroredCircleAndRotation", func(t *testing.T) {
		j := gjson.Parse(`{
			"RegionName":["Mirror","Circle","Rotated","Untyped"],
			"RegionType":["Face","Face","Face",""],
			"RegionAreaX":[0.2,0.5,0.7,0.5],
			"RegionAreaY":[0.3,0.5,0.5,0.5],
			"RegionAreaW":[0.1,0,0.2,0.2],
			"RegionAreaH":[0.2,0,0.2,0.2],
			"RegionAreaD":[0,0.2,0,0],
			"RegionRotation":[0,0,1.5707963,0]
		}`)

		faces := parseExiftoolFaces(j, FaceOptions{Orientation: 2, Width: 4000, Height: 2000})
		if len(faces) != 3 {
			t.Fatalf("want 3 typed faces, got %d: %+v", len(faces), faces)
		}
		if mirror := faces[0]; mirror.Name != "Mirror" || !almost(mirror.X, 0.75) || !almost(mirror.Y, 0.2) {
			t.Errorf("mirrored face got %+v", mirror)
		}
		if circle := faces[1]; circle.Name != "Circle" || !almost(circle.W, 0.1) || !almost(circle.H, 0.2) {
			t.Errorf("circle got %+v", circle)
		}
		if rotated := faces[2]; rotated.Name != "Rotated" || !almost(rotated.W, 0.1) || !almost(rotated.H, 0.4) {
			t.Errorf("rotated face got %+v", rotated)
		}
	})

	t.Run("ACDSee", func(t *testing.T) {
		j := gjson.Parse(`{
			"ACDSeeRegionName":["DLY","ALG","Object"],
			"ACDSeeRegionType":["Face","Face","Object"],
			"ACDSeeRegionDLYAreaX":[0.3],
			"ACDSeeRegionDLYAreaY":[0.4],
			"ACDSeeRegionDLYAreaW":[0.2],
			"ACDSeeRegionDLYAreaH":[0.3],
			"ACDSeeRegionALGAreaX":[0.8,0.7,0.5],
			"ACDSeeRegionALGAreaY":[0.8,0.6,0.5],
			"ACDSeeRegionALGAreaW":[0.1,0.1,0.2],
			"ACDSeeRegionALGAreaH":[0.1,0.2,0.2]
		}`)

		faces := parseExiftoolFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if len(faces) != 2 {
			t.Fatalf("want 2 ACDSee faces, got %d: %+v", len(faces), faces)
		}
		if faces[0].Name != "DLY" || !almost(faces[0].X, 0.2) || !almost(faces[0].Y, 0.25) {
			t.Errorf("DLY face got %+v", faces[0])
		}
		if faces[1].Name != "ALG" || !almost(faces[1].X, 0.65) || !almost(faces[1].Y, 0.5) {
			t.Errorf("ALG face got %+v", faces[1])
		}
	})

	t.Run("NameFallbacks", func(t *testing.T) {
		j := gjson.Parse(`{
			"RegionName":["","",""],
			"RegionTitle":["Title Name","",""],
			"RegionPersonInImage":["","Extension Name",""],
			"RegionSeeAlso":["","","Iptc4xmpExt:PersonInImage"],
			"PersonInImage":"Referenced Name",
			"RegionType":["Face","Face","Face"],
			"RegionAreaX":[0.2,0.5,0.8],
			"RegionAreaY":[0.5,0.5,0.5],
			"RegionAreaW":[0.1,0.1,0.1],
			"RegionAreaH":[0.1,0.1,0.1]
		}`)

		faces := parseExiftoolFaces(j, FaceOptions{Orientation: 1, Width: 4000, Height: 2000})
		if len(faces) != 3 {
			t.Fatalf("want 3 named faces, got %d: %+v", len(faces), faces)
		}
		if faces[0].Name != "Title Name" || faces[1].Name != "Extension Name" || faces[2].Name != "Referenced Name" {
			t.Errorf("unexpected name fallback result: %+v", faces)
		}
	})
}

func TestJsonAt(t *testing.T) {
	arr := gjson.Parse(`["a","b"]`).Array()
	if jsonAt(arr, 0).String() != "a" || jsonAt(arr, 1).String() != "b" {
		t.Error("in-range index must return the element")
	}
	if jsonAt(arr, 2).Exists() {
		t.Error("out-of-range index must return a non-existent result")
	}
	if jsonAt(arr, -1).Exists() {
		t.Error("negative index must return a non-existent result")
	}
}
