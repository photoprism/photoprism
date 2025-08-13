package tensorflow

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
)

var assetsPath = fs.Abs("../../../assets")
var testDataPath = fs.Abs("testdata")

func TestTF1ModelLoad(t *testing.T) {
	model, err := SavedModel(
		filepath.Join(assetsPath, "models", "nasnet"),
		[]string{"photoprism"})

	if err != nil {
		t.Fatal(err)
	}

	input, output, err := GetInputAndOutputFromSavedModel(model)
	if err == nil {
		t.Fatalf("TF1 does not have signatures, but GetInput worked")
	}

	input, output, err = GuessInputAndOutput(model)
	if err != nil {
		t.Fatal(err)
	}

	if input == nil {
		t.Fatal("Could not get the input")
	} else if output == nil {
		t.Fatal("Could not get the output")
	} else if input.Shape == nil {
		t.Fatal("Could not get the shape")
	} else {
		t.Logf("Shape: %v", input.Shape)
	}
}

func TestTF2ModelLoad(t *testing.T) {
	model, err := SavedModel(
		filepath.Join(testDataPath, "tf2"),
		[]string{"serve"})

	if err != nil {
		t.Fatal(err)
	}

	input, output, err := GetInputAndOutputFromSavedModel(model)
	if err != nil {
		t.Fatal(err)
	}

	if input == nil {
		t.Fatal("Could not get the input")
	} else if output == nil {
		t.Fatal("Could not get the output")
	} else if input.Shape == nil {
		t.Fatal("Could not get the shape")
	} else if !slices.Equal(input.Shape, DefaultPhotoInputShape()) {
		t.Fatalf("Invalid shape calculated. Expected BHWC, got %v",
			input.Shape)
	}
}

func TestTF2ModelBCHWLoad(t *testing.T) {
	model, err := SavedModel(
		filepath.Join(testDataPath, "tf2_bchw"),
		[]string{"serve"})

	if err != nil {
		t.Fatal(err)
	}

	input, output, err := GetInputAndOutputFromSavedModel(model)
	if err != nil {
		t.Fatal(err)
	}

	if input == nil {
		t.Fatal("Could not get the input")
	} else if output == nil {
		t.Fatal("Could not get the output")
	} else if input.Shape == nil {
		t.Fatal("Could not get the shape")
	} else if !slices.Equal(input.Shape, []ShapeComponent{
		ShapeBatch, ShapeColor, ShapeHeight, ShapeWidth,
	}) {
		t.Fatalf("Invalid shape calculated. Expected BCHW, got %v",
			input.Shape)
	}
}
