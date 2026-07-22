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
