package meta

import (
	"os"
	"testing"
)

func almost(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.001
}

func TestRotateRect(t *testing.T) {
	// Asymmetric rect near the top-left so every orientation lands distinctly.
	x, y, w, h := float32(0.1), float32(0.2), float32(0.3), float32(0.4)
	t.Run("Identity", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 1)
		if !almost(gx, 0.1) || !almost(gy, 0.2) || !almost(gw, 0.3) || !almost(gh, 0.4) {
			t.Errorf("got %v %v %v %v", gx, gy, gw, gh)
		}
	})
	t.Run("Zero", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 0)
		if !almost(gx, 0.1) || !almost(gy, 0.2) || !almost(gw, 0.3) || !almost(gh, 0.4) {
			t.Errorf("orientation 0 must be identity, got %v %v %v %v", gx, gy, gw, gh)
		}
	})
	t.Run("Rotate180", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 3)
		if !almost(gx, 0.6) || !almost(gy, 0.4) || !almost(gw, 0.3) || !almost(gh, 0.4) {
			t.Errorf("got %v %v %v %v, want 0.6 0.4 0.3 0.4", gx, gy, gw, gh)
		}
	})
	t.Run("MirrorHorizontal", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 2)
		if !almost(gx, 0.6) || !almost(gy, 0.2) || !almost(gw, 0.3) || !almost(gh, 0.4) {
			t.Errorf("got %v %v %v %v, want 0.6 0.2 0.3 0.4", gx, gy, gw, gh)
		}
	})
	t.Run("MirrorVertical", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 4)
		if !almost(gx, 0.1) || !almost(gy, 0.4) || !almost(gw, 0.3) || !almost(gh, 0.4) {
			t.Errorf("got %v %v %v %v, want 0.1 0.4 0.3 0.4", gx, gy, gw, gh)
		}
	})
	t.Run("Transpose", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 5)
		if !almost(gx, 0.2) || !almost(gy, 0.1) || !almost(gw, 0.4) || !almost(gh, 0.3) {
			t.Errorf("got %v %v %v %v, want 0.2 0.1 0.4 0.3", gx, gy, gw, gh)
		}
	})
	t.Run("Rotate90CW", func(t *testing.T) {
		// EXIF 6: x'=1-y-h, y'=x, w'=h, h'=w
		gx, gy, gw, gh := rotateRect(x, y, w, h, 6)
		if !almost(gx, 0.4) || !almost(gy, 0.1) || !almost(gw, 0.4) || !almost(gh, 0.3) {
			t.Errorf("got %v %v %v %v, want 0.4 0.1 0.4 0.3", gx, gy, gw, gh)
		}
	})
	t.Run("Rotate270CW", func(t *testing.T) {
		// EXIF 8: x'=y, y'=1-x-w, w'=h, h'=w
		gx, gy, gw, gh := rotateRect(x, y, w, h, 8)
		if !almost(gx, 0.2) || !almost(gy, 0.6) || !almost(gw, 0.4) || !almost(gh, 0.3) {
			t.Errorf("got %v %v %v %v, want 0.2 0.6 0.4 0.3", gx, gy, gw, gh)
		}
	})
	t.Run("Transverse", func(t *testing.T) {
		gx, gy, gw, gh := rotateRect(x, y, w, h, 7)
		if !almost(gx, 0.4) || !almost(gy, 0.6) || !almost(gw, 0.4) || !almost(gh, 0.3) {
			t.Errorf("got %v %v %v %v, want 0.4 0.6 0.4 0.3", gx, gy, gw, gh)
		}
	})
}

func TestNormalizeRegionMWG(t *testing.T) {
	t.Run("NormalizedCenterToTopLeft", func(t *testing.T) {
		// Center (0.5,0.5) size 0.2x0.3 -> top-left (0.4,0.35).
		f, ok := normalizeRegionMWG("Alice", 0.5, 0.5, 0.2, 0.3, 0, "normalized", 0, 0, 0, 1)
		if !ok {
			t.Fatal("expected valid region")
		}
		if f.Name != "Alice" || !almost(f.X, 0.4) || !almost(f.Y, 0.35) || !almost(f.W, 0.2) || !almost(f.H, 0.3) {
			t.Errorf("got %+v", f)
		}
	})
	t.Run("PixelWithAppliedDimensions", func(t *testing.T) {
		// Center at (2000,1500) size 400x600 against 4000x3000 -> center (0.5,0.5) size 0.1x0.2.
		f, ok := normalizeRegionMWG("Bob", 2000, 1500, 400, 600, 0, "pixel", 0, 4000, 3000, 1)
		if !ok {
			t.Fatal("expected valid region")
		}
		if !almost(f.X, 0.45) || !almost(f.Y, 0.4) || !almost(f.W, 0.1) || !almost(f.H, 0.2) {
			t.Errorf("got %+v", f)
		}
	})
	t.Run("PixelWithoutAppliedDimensionsRejected", func(t *testing.T) {
		if _, ok := normalizeRegionMWG("Bob", 2000, 1500, 400, 600, 0, "pixel", 0, 0, 0, 1); ok {
			t.Error("pixel unit without applied dimensions must be rejected")
		}
	})
	t.Run("NonPositiveRejected", func(t *testing.T) {
		if _, ok := normalizeRegionMWG("X", 0.5, 0.5, 0, 0.2, 0, "normalized", 0, 0, 0, 1); ok {
			t.Error("zero-width region must be rejected")
		}
	})
	t.Run("OrientationApplied", func(t *testing.T) {
		// Center (0.5,0.5) size 0.2x0.3 -> TL (0.4,0.35); rotate 90CW.
		f, ok := normalizeRegionMWG("Alice", 0.5, 0.5, 0.2, 0.3, 0, "normalized", 0, 0, 0, 6)
		if !ok {
			t.Fatal("expected valid region")
		}
		// rotateRect(0.4,0.35,0.2,0.3,6) = (1-0.35-0.3, 0.4, 0.3, 0.2) = (0.35,0.4,0.3,0.2)
		if !almost(f.X, 0.35) || !almost(f.Y, 0.4) || !almost(f.W, 0.3) || !almost(f.H, 0.2) {
			t.Errorf("got %+v", f)
		}
	})
	t.Run("ClipsLeftAndRightEdges", func(t *testing.T) {
		left, ok := normalizeRegionMWG("Left", 0.05, 0.5, 0.2, 0.2, 0, "normalized", 0, 0, 0, 1)
		if !ok || !almost(left.X, 0) || !almost(left.W, 0.15) {
			t.Errorf("left edge got %+v, valid %t", left, ok)
		}

		right, ok := normalizeRegionMWG("Right", 0.95, 0.5, 0.2, 0.2, 0, "normalized", 0, 0, 0, 1)
		if !ok || !almost(right.X, 0.85) || !almost(right.W, 0.15) {
			t.Errorf("right edge got %+v, valid %t", right, ok)
		}
	})
	t.Run("CenterOutsideRejected", func(t *testing.T) {
		if _, ok := normalizeRegionMWG("Outside", 1.1, 0.5, 0.2, 0.2, 0, "normalized", 0, 0, 0, 1); ok {
			t.Error("region with an out-of-range center must be rejected")
		}
	})
	t.Run("CircleUsesShortImageDimension", func(t *testing.T) {
		landscape, ok := normalizeRegionMWG("Circle", 0.5, 0.5, 0, 0, 0.2, "normalized", 0, 4000, 2000, 1)
		if !ok || !almost(landscape.X, 0.45) || !almost(landscape.Y, 0.4) || !almost(landscape.W, 0.1) || !almost(landscape.H, 0.2) {
			t.Errorf("landscape circle got %+v, valid %t", landscape, ok)
		}

		portrait, ok := normalizeRegionMWG("Circle", 0.5, 0.5, 0, 0, 0.2, "normalized", 0, 2000, 4000, 1)
		if !ok || !almost(portrait.X, 0.4) || !almost(portrait.Y, 0.45) || !almost(portrait.W, 0.2) || !almost(portrait.H, 0.1) {
			t.Errorf("portrait circle got %+v, valid %t", portrait, ok)
		}
	})
	t.Run("RotationUsesPixelAspectRatio", func(t *testing.T) {
		rotated, ok := normalizeRegionMWG("Rotated", 0.5, 0.5, 0.2, 0.2, 0, "normalized", 1.5707963, 4000, 2000, 1)
		if !ok || !almost(rotated.X, 0.45) || !almost(rotated.Y, 0.3) || !almost(rotated.W, 0.1) || !almost(rotated.H, 0.4) {
			t.Errorf("rotated region got %+v, valid %t", rotated, ok)
		}

		negative, ok := normalizeRegionMWG("Rotated", 0.5, 0.5, 0.2, 0.2, 0, "normalized", -1.5707963, 4000, 2000, 1)
		if !ok || !almost(negative.W, rotated.W) || !almost(negative.H, rotated.H) {
			t.Errorf("negative rotation got %+v, valid %t", negative, ok)
		}
	})
}

func TestNormalizeRegionMP(t *testing.T) {
	t.Run("TopLeftPassthrough", func(t *testing.T) {
		f, ok := normalizeRegionMP("Cara", "0.3, 0.2, 0.1, 0.15", 1)
		if !ok {
			t.Fatal("expected valid region")
		}
		if f.Name != "Cara" || !almost(f.X, 0.3) || !almost(f.Y, 0.2) || !almost(f.W, 0.1) || !almost(f.H, 0.15) {
			t.Errorf("got %+v", f)
		}
	})
	t.Run("MalformedRejected", func(t *testing.T) {
		if _, ok := normalizeRegionMP("Cara", "0.3, 0.2, 0.1", 1); ok {
			t.Error("three-value rectangle must be rejected")
		}
		if _, ok := normalizeRegionMP("Cara", "a, b, c, d", 1); ok {
			t.Error("non-numeric rectangle must be rejected")
		}
	})
	t.Run("UnnamedKept", func(t *testing.T) {
		f, ok := normalizeRegionMP("", "0.3, 0.2, 0.1, 0.15", 1)
		if !ok {
			t.Fatal("unnamed region with a valid rectangle must be kept")
		}
		if f.Name != "" {
			t.Errorf("want empty name, got %q", f.Name)
		}
	})
}

func TestDedupFaces(t *testing.T) {
	t.Run("ExactDuplicateSameNameMerged", func(t *testing.T) {
		in := []Face{
			{Name: "Alice", X: 0.4, Y: 0.35, W: 0.2, H: 0.3},
			{Name: "Alice", X: 0.4, Y: 0.35, W: 0.2, H: 0.3},
		}
		if out := DedupFaces(in); len(out) != 1 {
			t.Errorf("want 1, got %d", len(out))
		}
	})
	t.Run("DifferentPeopleKept", func(t *testing.T) {
		in := []Face{
			{Name: "Alice", X: 0.4, Y: 0.35, W: 0.2, H: 0.3},
			{Name: "Bob", X: 0.41, Y: 0.36, W: 0.2, H: 0.3},
		}
		if out := DedupFaces(in); len(out) != 2 {
			t.Errorf("overlapping different people must both be kept, got %d", len(out))
		}
	})
	t.Run("CaseInsensitiveName", func(t *testing.T) {
		in := []Face{
			{Name: "Alice", X: 0.4, Y: 0.35, W: 0.2, H: 0.3},
			{Name: "alice", X: 0.4, Y: 0.35, W: 0.2, H: 0.3},
		}
		if out := DedupFaces(in); len(out) != 1 {
			t.Errorf("same name differing only in case must merge, got %d", len(out))
		}
	})
}

const xmpFacesHeader = `<?xml version="1.0"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
 xmlns:tiff="http://ns.adobe.com/tiff/1.0/"
 xmlns:mwg-rs="http://www.metadataworkinggroup.com/schemas/regions/"
 xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#"
 xmlns:stArea="http://ns.adobe.com/xmp/sType/Area#"
 xmlns:stDim="http://ns.adobe.com/xmp/sType/Dimensions#"
 xmlns:Iptc4xmpExt="http://iptc.org/std/Iptc4xmpExt/2008-02-29/"
 xmlns:acdsee-rs="http://ns.acdsee.com/regions/"
 xmlns:acdsee-stArea="http://ns.acdsee.com/sType/Area#"
 xmlns:acdsee-stDim="http://ns.acdsee.com/sType/Dimensions#"
 xmlns:MP="http://ns.microsoft.com/photo/1.2/"
 xmlns:MPRI="http://ns.microsoft.com/photo/1.2/t/RegionInfo#"
 xmlns:MPReg="http://ns.microsoft.com/photo/1.2/t/Region#">
 <rdf:RDF><rdf:Description rdf:about="">`

const xmpFacesFooter = `</rdf:Description></rdf:RDF></x:xmpmeta>`

func TestXmpDocument_Faces_MWG(t *testing.T) {
	body := xmpFacesHeader + `
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:AppliedToDimensions stDim:w="4000" stDim:h="3000" stDim:unit="pixel"/>
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Alice</mwg-rs:Name>
     <mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.5" stArea:y="0.5" stArea:w="0.2" stArea:h="0.3" stArea:unit="normalized"/>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 1 {
		t.Fatalf("want 1 face, got %d", len(faces))
	}
	f := faces[0]
	if f.Name != "Alice" || !almost(f.X, 0.4) || !almost(f.Y, 0.35) || !almost(f.W, 0.2) || !almost(f.H, 0.3) {
		t.Errorf("got %+v", f)
	}
}

// TestXmpDocument_Faces_LightroomNestedDescription verifies MWG attributes on nested RDF descriptions.
func TestXmpDocument_Faces_LightroomNestedDescription(t *testing.T) {
	body := xmpFacesHeader + `
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:AppliedToDimensions stDim:w="6000" stDim:h="4000" stDim:unit="pixel"/>
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li>
     <rdf:Description mwg-rs:Rotation="0.00000" mwg-rs:Name="Anja" mwg-rs:Type="Face">
      <mwg-rs:Area stArea:x="0.39404" stArea:y="0.34384" stArea:w="0.04199" stArea:h="0.06305"/>
     </rdf:Description>
    </rdf:li>
    <rdf:li>
     <rdf:Description mwg-rs:Rotation="0.00000" mwg-rs:Name="Barcode123" mwg-rs:Type="BarCode">
      <mwg-rs:Area stArea:x="0.1" stArea:y="0.1" stArea:w="0.05" stArea:h="0.05"/>
     </rdf:Description>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 1 {
		t.Fatalf("want 1 face, got %d", len(faces))
	}
	f := faces[0]
	if f.Name != "Anja" || !almost(f.X, 0.373045) || !almost(f.Y, 0.312315) || !almost(f.W, 0.04199) || !almost(f.H, 0.06305) {
		t.Errorf("got %+v", f)
	}
}

func TestXmpDocument_Faces_DigikamPixel(t *testing.T) {
	// digiKam-style: pixel area units resolved against AppliedToDimensions.
	body := xmpFacesHeader + `
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:AppliedToDimensions stDim:w="4000" stDim:h="3000"/>
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Carl</mwg-rs:Name>
     <mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="2000" stArea:y="1500" stArea:w="400" stArea:h="600" stArea:unit="pixel"/>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 1 {
		t.Fatalf("want 1 face, got %d", len(faces))
	}
	f := faces[0]
	if f.Name != "Carl" || !almost(f.X, 0.45) || !almost(f.Y, 0.4) || !almost(f.W, 0.1) || !almost(f.H, 0.2) {
		t.Errorf("got %+v", f)
	}
}

// TestXmpDocument_Faces_MWGEdgeCases verifies optional MWG geometry and names.
func TestXmpDocument_Faces_MWGEdgeCases(t *testing.T) {
	body := xmpFacesHeader + `
  <Iptc4xmpExt:PersonInImage><rdf:Bag><rdf:li>Referenced Person</rdf:li></rdf:Bag></Iptc4xmpExt:PersonInImage>
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:AppliedToDimensions stDim:w="4000" stDim:h="2000" stDim:unit="pixel"/>
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Title>Title Person</mwg-rs:Title><mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.2" stArea:y="0.2" stArea:w="0.1" stArea:h="0.2"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Type>Face</mwg-rs:Type><rdfs:seeAlso rdf:resource="Iptc4xmpExt:PersonInImage"/>
     <mwg-rs:Area stArea:x="0.4" stArea:y="0.2" stArea:w="0.1" stArea:h="0.2"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Extensions><Iptc4xmpExt:PersonInImage>Extension Person</Iptc4xmpExt:PersonInImage></mwg-rs:Extensions>
     <mwg-rs:Area stArea:x="0.6" stArea:y="0.2" stArea:w="0.1" stArea:h="0.2"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Circle Person</mwg-rs:Name><mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.5" stArea:y="0.5" stArea:d="0.2" stArea:unit="normalized"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Rotated Person</mwg-rs:Name><mwg-rs:Type>Face</mwg-rs:Type><mwg-rs:Rotation>1.5707963</mwg-rs:Rotation>
     <mwg-rs:Area stArea:x="0.8" stArea:y="0.6" stArea:w="0.2" stArea:h="0.2"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Invalid Rotation</mwg-rs:Name><mwg-rs:Type>Face</mwg-rs:Type><mwg-rs:Rotation>invalid</mwg-rs:Rotation>
     <mwg-rs:Area stArea:x="0.2" stArea:y="0.8" stArea:w="0.1" stArea:h="0.1"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Untyped Object</mwg-rs:Name>
     <mwg-rs:Area stArea:x="0.9" stArea:y="0.9" stArea:w="0.1" stArea:h="0.1"/>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 6 {
		t.Fatalf("want 6 typed faces, got %d: %+v", len(faces), faces)
	}

	byName := make(map[string]Face, len(faces))
	for _, parsed := range faces {
		byName[parsed.Name] = parsed
	}

	for _, name := range []string{"Title Person", "Referenced Person", "Extension Person", "Circle Person", "Rotated Person", "Invalid Rotation"} {
		if _, exists := byName[name]; !exists {
			t.Errorf("missing resolved face name %q: %+v", name, faces)
		}
	}
	if circle := byName["Circle Person"]; !almost(circle.W, 0.1) || !almost(circle.H, 0.2) {
		t.Errorf("circle got %+v", circle)
	}
	if rotated := byName["Rotated Person"]; !almost(rotated.W, 0.1) || !almost(rotated.H, 0.4) {
		t.Errorf("rotated rectangle got %+v", rotated)
	}
	if invalid := byName["Invalid Rotation"]; !almost(invalid.W, 0.1) || !almost(invalid.H, 0.1) {
		t.Errorf("invalid rotation must be treated as zero, got %+v", invalid)
	}
}

// TestXmpDocument_Faces_ACDSee verifies DLY and ALG face regions.
func TestXmpDocument_Faces_ACDSee(t *testing.T) {
	body := xmpFacesHeader + `
  <acdsee-rs:Regions rdf:parseType="Resource">
   <acdsee-rs:AppliedToDimensions acdsee-stDim:w="4000" acdsee-stDim:h="2000" acdsee-stDim:unit="pixel"/>
   <acdsee-rs:RegionList><rdf:Seq>
    <rdf:li rdf:parseType="Resource">
     <acdsee-rs:Name>DLY Person</acdsee-rs:Name><acdsee-rs:Type>Face</acdsee-rs:Type>
     <acdsee-rs:DLYArea acdsee-stArea:x="0.3" acdsee-stArea:y="0.4" acdsee-stArea:w="0.2" acdsee-stArea:h="0.3"/>
     <acdsee-rs:ALGArea acdsee-stArea:x="0.8" acdsee-stArea:y="0.8" acdsee-stArea:w="0.1" acdsee-stArea:h="0.1"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <acdsee-rs:Name>ALG Person</acdsee-rs:Name><acdsee-rs:Type>Face</acdsee-rs:Type>
     <acdsee-rs:ALGArea acdsee-stArea:x="0.7" acdsee-stArea:y="0.6" acdsee-stArea:w="0.1" acdsee-stArea:h="0.2"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <acdsee-rs:Name>Chair</acdsee-rs:Name><acdsee-rs:Type>Object</acdsee-rs:Type>
     <acdsee-rs:DLYArea acdsee-stArea:x="0.5" acdsee-stArea:y="0.5" acdsee-stArea:w="0.2" acdsee-stArea:h="0.2"/>
    </rdf:li>
   </rdf:Seq></acdsee-rs:RegionList>
  </acdsee-rs:Regions>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 2 {
		t.Fatalf("want 2 ACDSee faces, got %d: %+v", len(faces), faces)
	}
	if faces[0].Name != "DLY Person" || !almost(faces[0].X, 0.2) || !almost(faces[0].Y, 0.25) {
		t.Errorf("DLY region got %+v", faces[0])
	}
	if faces[1].Name != "ALG Person" || !almost(faces[1].X, 0.65) || !almost(faces[1].Y, 0.5) {
		t.Errorf("ALG region got %+v", faces[1])
	}
}

func TestXmpDocument_Faces_MP(t *testing.T) {
	body := xmpFacesHeader + `
  <MP:RegionInfo rdf:parseType="Resource"><MPRI:Regions><rdf:Bag>
   <rdf:li MPReg:Rectangle="0.3, 0.2, 0.1, 0.15" MPReg:PersonDisplayName="Cara"/>
  </rdf:Bag></MPRI:Regions></MP:RegionInfo>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 1 {
		t.Fatalf("want 1 face, got %d", len(faces))
	}
	f := faces[0]
	if f.Name != "Cara" || !almost(f.X, 0.3) || !almost(f.Y, 0.2) || !almost(f.W, 0.1) || !almost(f.H, 0.15) {
		t.Errorf("got %+v", f)
	}
}

func TestXmpDocument_Faces_UnnamedAndNonFace(t *testing.T) {
	body := xmpFacesHeader + `
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.5" stArea:y="0.5" stArea:w="0.2" stArea:h="0.2"/>
    </rdf:li>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Barcode123</mwg-rs:Name>
     <mwg-rs:Type>BarCode</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.1" stArea:y="0.1" stArea:w="0.1" stArea:h="0.1"/>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	faces := loadXmpString(t, body).Faces(1)
	if len(faces) != 1 {
		t.Fatalf("want 1 face (unnamed kept, barcode skipped), got %d", len(faces))
	}
	if faces[0].Name != "" {
		t.Errorf("want unnamed face, got %q", faces[0].Name)
	}
}

func TestXmpDocument_Faces_None(t *testing.T) {
	if faces := loadXmpString(t, xmpFacesHeader+xmpFacesFooter).Faces(1); len(faces) != 0 {
		t.Errorf("want 0 faces, got %d", len(faces))
	}
}

func TestFace_Valid(t *testing.T) {
	assert := func(f Face, want bool) {
		if f.Valid() != want {
			t.Errorf("Face%+v Valid()=%v, want %v", f, f.Valid(), want)
		}
	}
	assert(Face{X: 0.1, Y: 0.1, W: 0.2, H: 0.2}, true)
	assert(Face{X: 0.1, Y: 0.1, W: 0, H: 0.2}, false)    // zero width
	assert(Face{X: -0.1, Y: 0.1, W: 0.2, H: 0.2}, false) // negative origin
	assert(Face{X: 0.9, Y: 0.1, W: 0.2, H: 0.2}, false)  // exceeds right edge
}

func TestXmpDocument_Orientation(t *testing.T) {
	body := xmpFacesHeader + `<tiff:Orientation>6</tiff:Orientation>` + xmpFacesFooter
	if got := loadXmpString(t, body).Orientation(); got != 6 {
		t.Errorf("orientation = %d, want 6", got)
	}
	if got := loadXmpString(t, xmpFacesHeader+xmpFacesFooter).Orientation(); got != 0 {
		t.Errorf("missing orientation = %d, want 0", got)
	}
}

func TestXMP_Faces_RotatedSidecar(t *testing.T) {
	// A sidecar carrying tiff:Orientation must rotate its region into the
	// displayed frame end-to-end via meta.XMP.
	body := xmpFacesHeader + `<tiff:Orientation>6</tiff:Orientation>
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Rita</mwg-rs:Name>
     <mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.5" stArea:y="0.4" stArea:w="0.1" stArea:h="0.15" stArea:unit="normalized"/>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	tmp := t.TempDir() + "/rotated.xmp"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil { //nolint:gosec // test writes a temporary fixture
		t.Fatal(err)
	}
	data, err := XMP(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Faces) != 1 {
		t.Fatalf("want 1 face, got %d", len(data.Faces))
	}
	f := data.Faces[0]
	// center (0.5,0.4) size (0.1,0.15) -> TL (0.45,0.325); rotateRect(...,6) = (0.525,0.45,0.15,0.1).
	if f.Name != "Rita" || !almost(f.X, 0.525) || !almost(f.Y, 0.45) || !almost(f.W, 0.15) || !almost(f.H, 0.10) {
		t.Errorf("got %+v", f)
	}
}

// TestXMPWithOptions_FacesFallbacks verifies dimensions and orientation from the source image.
func TestXMPWithOptions_FacesFallbacks(t *testing.T) {
	body := xmpFacesHeader + `
  <mwg-rs:Regions rdf:parseType="Resource">
   <mwg-rs:RegionList><rdf:Bag>
    <rdf:li rdf:parseType="Resource">
     <mwg-rs:Name>Fallback</mwg-rs:Name><mwg-rs:Type>Face</mwg-rs:Type>
     <mwg-rs:Area stArea:x="0.25" stArea:y="0.5" stArea:d="0.2" stArea:unit="normalized"/>
    </rdf:li>
   </rdf:Bag></mwg-rs:RegionList>
  </mwg-rs:Regions>` + xmpFacesFooter

	tmp := t.TempDir() + "/fallback.xmp"
	if err := os.WriteFile(tmp, []byte(body), 0o644); err != nil { //nolint:gosec // test writes a temporary fixture
		t.Fatal(err)
	}
	data, err := XMPWithOptions(tmp, FaceOptions{Orientation: 2, Width: 4000, Height: 2000})
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Faces) != 1 {
		t.Fatalf("want 1 face, got %d", len(data.Faces))
	}
	f := data.Faces[0]
	if f.Name != "Fallback" || !almost(f.X, 0.7) || !almost(f.Y, 0.4) || !almost(f.W, 0.1) || !almost(f.H, 0.2) {
		t.Errorf("got %+v", f)
	}
}

func TestXMP_Faces_SidecarFixtures(t *testing.T) {
	// Exercises the real .xmp sidecar fixtures end-to-end through meta.XMP, one
	// per supported writer/schema, so the Adobe/Lightroom, digiKam, and Microsoft
	// region formats are each covered by an actual file (not only inline XMP).
	// The expected values are displayed-orientation top-left rectangles.
	cases := []struct {
		name string
		path string
		want Face
	}{
		{
			// center (0.4,0.35) size (0.12,0.16) -> top-left (0.34,0.27).
			name: "AdobeLightroomMWG",
			path: "testdata/faces/adobe-mwg.xmp",
			want: Face{Name: "Diana", X: 0.34, Y: 0.27, W: 0.12, H: 0.16},
		},
		{
			// center (0.6,0.5) size (0.1,0.15) -> top-left (0.55,0.425).
			name: "DigikamMWG",
			path: "testdata/faces/digikam-mwg.xmp",
			want: Face{Name: "Carl", X: 0.55, Y: 0.425, W: 0.1, H: 0.15},
		},
		{
			name: "MicrosoftMP",
			path: "testdata/faces/microsoft-mp.xmp",
			want: Face{Name: "Cara", X: 0.3, Y: 0.2, W: 0.1, H: 0.15},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := XMP(tc.path)
			if err != nil {
				t.Fatalf("XMP(%s): %v", tc.path, err)
			}
			if len(data.Faces) != 1 {
				t.Fatalf("want 1 face, got %d (%+v)", len(data.Faces), data.Faces)
			}
			f := data.Faces[0]
			if f.Name != tc.want.Name || !almost(f.X, tc.want.X) || !almost(f.Y, tc.want.Y) || !almost(f.W, tc.want.W) || !almost(f.H, tc.want.H) {
				t.Errorf("got %+v, want %+v", f, tc.want)
			}
		})
	}
}

func TestParseFloat32(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		v, ok := parseFloat32(" 0.5 ")
		if !ok || !almost(v, 0.5) {
			t.Errorf("got %v %v", v, ok)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		if _, ok := parseFloat32("  "); ok {
			t.Error("empty must fail")
		}
	})
	t.Run("NonNumeric", func(t *testing.T) {
		if _, ok := parseFloat32("abc"); ok {
			t.Error("non-numeric must fail")
		}
	})
}

func TestClampUnit(t *testing.T) {
	if clampUnit(-0.1) != 0 {
		t.Error("negative must clamp to 0")
	}
	if clampUnit(1.5) != 1 {
		t.Error(">1 must clamp to 1")
	}
	if !almost(clampUnit(0.3), 0.3) {
		t.Error("in-range must pass through")
	}
}
