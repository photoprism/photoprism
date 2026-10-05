package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType_String(t *testing.T) {
	t.Run("Jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", ImageJpeg.String())
	})
}

func TestType_ToUpper(t *testing.T) {
	assert.Equal(t, "JPG", ImageJpeg.ToUpper())
}

func TestType_Equal(t *testing.T) {
	t.Run("Jpg", func(t *testing.T) {
		assert.True(t, ImageJpeg.Equal("jpg"))
	})
}

func TestType_NotEqual(t *testing.T) {
	t.Run("Jpg", func(t *testing.T) {
		assert.False(t, ImageJpeg.NotEqual("JPG"))
		assert.True(t, ImageJpeg.NotEqual("xmp"))
	})
}

func TestType_DefaultExt(t *testing.T) {
	t.Run("Jpg", func(t *testing.T) {
		assert.Equal(t, ".jpg", ImageJpeg.DefaultExt())
	})
	t.Run("Avif", func(t *testing.T) {
		assert.Equal(t, ".avif", ImageAvif.DefaultExt())
	})
}

func TestToType(t *testing.T) {
	t.Run("Jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", NewType("JPG").String())
	})
	t.Run("JPEG", func(t *testing.T) {
		assert.Equal(t, Type("jpeg"), NewType("JPEG"))
	})
	t.Run("Jpg", func(t *testing.T) {
		assert.Equal(t, "jpg", NewType(".jpg").String())
	})
}

func TestType_Is(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, ImageJpeg.Equal(""))
	})
	t.Run("Upper", func(t *testing.T) {
		assert.True(t, ImageJpeg.Equal("JPG"))
	})
	t.Run("Lower", func(t *testing.T) {
		assert.True(t, ImageJpeg.Equal("jpg"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, ImageJpeg.Equal("raw"))
	})
}

func TestType_Find(t *testing.T) {
	t.Run("FindJpg", func(t *testing.T) {
		result := ImageJpeg.Find("testdata/test.xmp", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("UpperExt", func(t *testing.T) {
		result := ImageJpeg.Find("testdata/test.XMP", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("WithSequence", func(t *testing.T) {
		result := ImageJpeg.Find("testdata/test (2).xmp", false)
		assert.Equal(t, "", result)
	})
	t.Run("StripSequence", func(t *testing.T) {
		result := ImageJpeg.Find("testdata/test (2).xmp", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("NameUpper", func(t *testing.T) {
		result := ImageJpeg.Find("testdata/CATYELLOW.xmp", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})
	t.Run("NameLower", func(t *testing.T) {
		result := ImageJpeg.Find("testdata/chameleon_lime.xmp", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
}

func TestType_FindFirst(t *testing.T) {
	dirs := []string{PPHiddenPathname}

	t.Run("FindXmp", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test.jpg", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})
	t.Run("FindXmpUpperExt", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test.PNG", dirs, "", false)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})
	t.Run("FindXmpWithoutSequence", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test (2).jpg", dirs, "", false)
		assert.Equal(t, "", result)
	})
	t.Run("FindXmpWithSequence", func(t *testing.T) {
		result := SidecarXMP.FindFirst("testdata/test (2).jpg", dirs, "", true)
		assert.Equal(t, "testdata/.photoprism/test.xmp", result)
	})
	t.Run("FindJpg", func(t *testing.T) {
		result := ImageJpeg.FindFirst("testdata/test.xmp", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("FindJpgAbs", func(t *testing.T) {
		result := ImageJpeg.FindFirst(Abs("testdata/test.xmp"), dirs, "", false)
		assert.Equal(t, Abs("testdata/test.jpg"), result)
	})
	t.Run("UpperExt", func(t *testing.T) {
		result := ImageJpeg.FindFirst("testdata/test.XMP", dirs, "", false)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("WithSequence", func(t *testing.T) {
		result := ImageJpeg.FindFirst("testdata/test (2).xmp", dirs, "", false)
		assert.Equal(t, "", result)
	})
	t.Run("StripSequence", func(t *testing.T) {
		result := ImageJpeg.FindFirst("testdata/test (2).xmp", dirs, "", true)
		assert.Equal(t, "testdata/test.jpg", result)
	})
	t.Run("NameUpper", func(t *testing.T) {
		result := ImageJpeg.FindFirst("testdata/CATYELLOW.xmp", dirs, "", true)
		assert.Equal(t, "testdata/CATYELLOW.jpg", result)
	})
	t.Run("NameLower", func(t *testing.T) {
		result := ImageJpeg.FindFirst("testdata/chameleon_lime.xmp", dirs, "", true)
		assert.Equal(t, "testdata/chameleon_lime.jpg", result)
	})
	t.Run("ExampleBmpNotfound", func(t *testing.T) {
		result := ImageBmp.FindFirst("testdata/example.00001.jpg", dirs, "", true)
		assert.Equal(t, "", result)
	})
	t.Run("ExampleBmpFound", func(t *testing.T) {
		result := ImageBmp.FindFirst("testdata/example.00001.jpg", []string{"directory"}, "", true)
		assert.Equal(t, "testdata/directory/example.bmp", result)
	})
	t.Run("ExamplePngFound", func(t *testing.T) {
		result := ImagePng.FindFirst("testdata/example.00001.jpg", []string{"directory", "directory/subdirectory"}, "", true)
		assert.Equal(t, "testdata/directory/subdirectory/example.png", result)
	})
	t.Run("ExampleBmpFound", func(t *testing.T) {
		result := ImageBmp.FindFirst(Abs("testdata/example.00001.jpg"), []string{"directory"}, Abs("testdata"), true)
		assert.Equal(t, Abs("testdata/directory/example.bmp"), result)
	})
}

func TestType_FindAll(t *testing.T) {
	dirs := []string{PPHiddenPathname}

	t.Run("CatyellowJpg", func(t *testing.T) {
		result := ImageJpeg.FindAll("testdata/CATYELLOW.JSON", dirs, "", false)
		assert.Contains(t, result, "testdata/CATYELLOW.jpg")
	})
	t.Run("EmptyBase", func(t *testing.T) {
		// A folder name has no base and finds no files.
		parent := t.TempDir()
		dir := filepath.Join(parent, "2024")
		require.NoError(t, os.MkdirAll(dir, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(parent, "2024.jpg"), []byte("x"), 0o600))
		for _, name := range []string{dir + string(os.PathSeparator) + ".", dir + string(os.PathSeparator) + ".."} {
			assert.Empty(t, ImageJpeg.FindAll(name, dirs, parent, false), name)
			assert.Equal(t, "", ImageJpeg.FindFirst(name, dirs, parent, false), name)
			assert.Equal(t, "", ImageJpeg.Find(name, false), name)
		}
	})
	t.Run("DotName", func(t *testing.T) {
		// Names whose base is or reduces to a dot name find the files named after their own base.
		dir := t.TempDir()
		for name, want := range map[string]string{"..heic": "..heic.jpg", "...heic": "...heic.jpg", " ..heic": " ..jpg", " ...heic": " ...jpg"} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, want), []byte("x"), 0o600))
			assert.Equal(t, filepath.Join(dir, want), ImageJpeg.Find(filepath.Join(dir, name), true), name)
			assert.Equal(t, filepath.Join(dir, want), ImageJpeg.FindFirst(filepath.Join(dir, name), dirs, dir, true), name)
			assert.Contains(t, ImageJpeg.FindAll(filepath.Join(dir, name), dirs, dir, true), filepath.Join(dir, want), name)
		}
	})
	t.Run("SequenceOnly", func(t *testing.T) {
		// A sequence-only name keeps its own base when sequences are stripped.
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "(1).jpg"), []byte("x"), 0o600))
		assert.Equal(t, filepath.Join(dir, "(1).jpg"), ImageJpeg.FindFirst(filepath.Join(dir, "(1).heic"), dirs, dir, true))
		assert.Contains(t, ImageJpeg.FindAll(filepath.Join(dir, "(1).heic"), dirs, dir, true), filepath.Join(dir, "(1).jpg"))
	})
	t.Run("SiblingRoot", func(t *testing.T) {
		// Sidecars are looked up only below the base folder.
		base := t.TempDir()
		originals := filepath.Join(base, "photos")
		sidecar := filepath.Join(base, "sidecar")
		require.NoError(t, os.MkdirAll(filepath.Join(sidecar, "2024"), 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(sidecar, "2024", "IMG_1.heic.jpg"), []byte("x"), 0o600))
		fileName := filepath.Join(base, "photos2", "2024", "IMG_1.heic")
		assert.Empty(t, ImageJpeg.FindAll(fileName, []string{sidecar}, originals, false))
		assert.Equal(t, "", ImageJpeg.FindFirst(fileName, []string{sidecar}, originals, false))
		assert.Equal(t, filepath.Join(sidecar, "2024", "IMG_1.heic.jpg"), ImageJpeg.FindFirst(filepath.Join(originals, "2024", "IMG_1.heic"), []string{sidecar}, originals, false))
	})
}

func TestType_FindEach(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "originals", "2024")
	sidecar := filepath.Join(parent, "sidecar")
	dirs := []string{sidecar, PPHiddenPathname}
	fileName := filepath.Join(dir, "IMG_1.heic")
	files := []string{
		filepath.Join(dir, "IMG_1.jpg"),
		filepath.Join(sidecar, "2024", "IMG_1.heic.jpg"),
		filepath.Join(dir, PPHiddenPathname, "IMG_1.jpg"),
	}

	for _, name := range files {
		require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o700))
		require.NoError(t, os.WriteFile(name, []byte("x"), 0o600))
	}

	find := func(accept func(string) bool) string {
		return ImageJpeg.FindEach(fileName, dirs, filepath.Join(parent, "originals"), false, accept)
	}

	t.Run("NilAccept", func(t *testing.T) {
		assert.Equal(t, files[0], find(nil))
		assert.Equal(t, ImageJpeg.FindFirst(fileName, dirs, filepath.Join(parent, "originals"), false), find(nil))
	})
	t.Run("StopsAtFirstAccepted", func(t *testing.T) {
		var visited []string
		assert.Equal(t, files[0], find(func(name string) bool { visited = append(visited, name); return true }))
		assert.Equal(t, []string{files[0]}, visited)
	})
	t.Run("SkipsRejected", func(t *testing.T) {
		assert.Equal(t, files[1], find(func(name string) bool { return name != files[0] }))
		assert.Equal(t, files[2], find(func(name string) bool { return name == files[2] }))
	})
	t.Run("FindAllOrder", func(t *testing.T) {
		var visited []string
		assert.Equal(t, "", find(func(name string) bool { visited = append(visited, name); return false }))
		assert.Equal(t, ImageJpeg.FindAll(fileName, dirs, filepath.Join(parent, "originals"), false), visited)
		assert.Equal(t, []string{files[0], files[1], files[2]}, visited)
	})
	t.Run("NameOrder", func(t *testing.T) {
		// The full name comes before the base name, and both before their lower and upper case variants.
		dir := t.TempDir()
		for _, name := range []string{"Img_2.heic.jpg", "Img_2.jpg", "img_2.jpg", "IMG_2.jpg"} {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600))
		}
		var visited []string
		ImageJpeg.FindEach(filepath.Join(dir, "Img_2.heic"), nil, dir, false, func(name string) bool { visited = append(visited, filepath.Base(name)); return false })
		assert.Equal(t, []string{"Img_2.heic.jpg", "Img_2.jpg", "img_2.jpg", "IMG_2.jpg"}, visited)
	})
	t.Run("ExtensionOrder", func(t *testing.T) {
		// Each extension is searched in all directories before the next one.
		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_3.jpeg"), []byte("x"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(sidecar, "2024", "IMG_3.heic.jpg"), []byte("x"), 0o600))
		assert.Equal(t, filepath.Join(sidecar, "2024", "IMG_3.heic.jpg"), ImageJpeg.FindFirst(filepath.Join(dir, "IMG_3.heic"), dirs, filepath.Join(parent, "originals"), false))
	})
	t.Run("RepeatedDir", func(t *testing.T) {
		// Each directory is checked once, also if it is given again or resolves to the folder of the file.
		repeated := []string{sidecar, dir, sidecar, ".", PPHiddenPathname, sidecar, PPHiddenPathname}
		assert.Equal(t, files, ImageJpeg.FindAll(fileName, repeated, filepath.Join(parent, "originals"), false))
	})
	t.Run("EmptyBase", func(t *testing.T) {
		called := false
		assert.Equal(t, "", ImageJpeg.FindEach(dir+string(os.PathSeparator)+".", dirs, parent, false, func(string) bool { called = true; return true }))
		assert.False(t, called)
	})
}

func TestFullNameCases(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, []string{"IMG_1234.raw", "IMG_1234.RAW", "img_1234.raw", "img_1234.RAW"}, fullNameCases("/photos/IMG_1234.raw"))
		assert.Equal(t, []string{"Img_1.Cr2", "Img_1.cr2", "Img_1.CR2", "img_1.Cr2", "img_1.cr2", "img_1.CR2", "IMG_1.Cr2", "IMG_1.cr2", "IMG_1.CR2"}, fullNameCases("Img_1.Cr2"))
	})
	t.Run("NoExtension", func(t *testing.T) {
		assert.Equal(t, []string{"IMG_1", "img_1"}, fullNameCases("/photos/IMG_1"))
	})
	t.Run("DotName", func(t *testing.T) {
		assert.Equal(t, []string{"..jpg", "..JPG"}, fullNameCases("/photos/..jpg"))
	})
}

func TestType_FindGenerated_SidecarCases(t *testing.T) {
	// Each of these names is found for a RAW file named IMG_1234.raw or img_1234.RAW, in any case.
	names := []string{
		"img_1234.jpg", "IMG_1234.jpg", "img_1234.JPG", "IMG_1234.JPG",
		"img_1234.jpeg", "IMG_1234.jpeg", "img_1234.JPEG", "IMG_1234.JPEG",
		"img_1234.raw.jpg", "IMG_1234.raw.jpg", "img_1234.RAW.JPG", "IMG_1234.RAW.JPG",
		"img_1234.raw.jpeg", "IMG_1234.raw.jpeg", "img_1234.RAW.JPEG", "IMG_1234.RAW.JPEG",
	}

	dir := t.TempDir()

	for _, name := range names {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600))
	}

	for _, original := range []string{"IMG_1234.raw", "img_1234.RAW", "IMG_1234.RAW", "img_1234.raw"} {
		var found []string
		ImageJpeg.FindGenerated(filepath.Join(dir, original), nil, dir, false, func(name string) bool {
			found = append(found, filepath.Base(name))
			return false
		})
		assert.ElementsMatch(t, names, found, original)
	}

	// Other lookups only check the full name as given.
	assert.NotContains(t, ImageJpeg.FindAll(filepath.Join(dir, "IMG_1234.raw"), nil, dir, false), filepath.Join(dir, "img_1234.RAW.JPG"))
}

func TestType_FindGenerated(t *testing.T) {
	parent := t.TempDir()
	originals := filepath.Join(parent, "originals")
	dir := filepath.Join(originals, "2024")
	sidecar := filepath.Join(parent, "sidecar")
	dirs := []string{sidecar, PPHiddenPathname}

	// write creates empty files with the given names.
	write := func(t *testing.T, names ...string) {
		for _, name := range names {
			require.NoError(t, os.MkdirAll(filepath.Dir(name), 0o700))
			require.NoError(t, os.WriteFile(name, []byte("x"), 0o600))
		}
	}

	// findAll returns the names FindGenerated visits for the given original.
	findAll := func(fileName string) (visited []string) {
		ImageJpeg.FindGenerated(fileName, dirs, originals, false, func(name string) bool {
			visited = append(visited, name)
			return false
		})
		return visited
	}

	t.Run("OwnFolder", func(t *testing.T) {
		write(t, filepath.Join(dir, "IMG_1.jpeg"), filepath.Join(dir, "img_1.JPG"))
		assert.ElementsMatch(t, []string{filepath.Join(dir, "IMG_1.jpeg"), filepath.Join(dir, "img_1.JPG")}, findAll(filepath.Join(dir, "IMG_1.heic")))
	})
	t.Run("FullNameCasesLast", func(t *testing.T) {
		// The full name in other cases comes after the other names of the same extension.
		write(t, filepath.Join(dir, "img_5.HEIC.jpg"), filepath.Join(dir, "IMG_5.jpg"))
		assert.Equal(t, []string{filepath.Join(dir, "IMG_5.jpg"), filepath.Join(dir, "img_5.HEIC.jpg")}, findAll(filepath.Join(dir, "IMG_5.heic")))
	})
	t.Run("GeneratedNames", func(t *testing.T) {
		generated := []string{
			filepath.Join(sidecar, "2024", "IMG_2.heic.jpg"),
			filepath.Join(sidecar, "2024", "IMG_2.jpg"),
			filepath.Join(dir, PPHiddenPathname, "IMG_2.heic.jpg"),
		}
		write(t, generated...)
		write(t,
			filepath.Join(sidecar, "2024", "IMG_2.HEIC.jpg"),
			filepath.Join(sidecar, "2024", "img_2.heic.jpg"),
			filepath.Join(sidecar, "2024", "IMG_2.heic.JPG"),
			filepath.Join(sidecar, "2024", "IMG_2.jpeg"),
			filepath.Join(sidecar, "2024", "img_2.jpg"),
			filepath.Join(dir, PPHiddenPathname, "IMG_2.JPG"),
		)
		assert.Equal(t, generated, findAll(filepath.Join(dir, "IMG_2.heic")))
		assert.Equal(t, generated[0], ImageJpeg.FindGenerated(filepath.Join(dir, "IMG_2.heic"), dirs, originals, false, nil))
	})
	t.Run("RelativeSidecar", func(t *testing.T) {
		write(t, filepath.Join(dir, ".sidecar", "IMG_3.heic.jpg"), filepath.Join(dir, ".sidecar", "IMG_3.JPEG"))
		var visited []string
		ImageJpeg.FindGenerated(filepath.Join(dir, "IMG_3.heic"), []string{".sidecar"}, originals, false, func(name string) bool { visited = append(visited, name); return false })
		assert.Equal(t, []string{filepath.Join(dir, ".sidecar", "IMG_3.heic.jpg")}, visited)
	})
	t.Run("SidecarInOriginals", func(t *testing.T) {
		// A sidecar folder that is the folder of the file keeps all variants.
		write(t, filepath.Join(dir, "IMG_4.heic.jpg"), filepath.Join(dir, "IMG_4.JPG"))
		want := []string{filepath.Join(dir, "IMG_4.heic.jpg"), filepath.Join(dir, "IMG_4.JPG")}
		for _, sidecarDir := range []string{".", originals} {
			var visited []string
			ImageJpeg.FindGenerated(filepath.Join(dir, "IMG_4.heic"), []string{sidecarDir, PPHiddenPathname}, originals, false, func(name string) bool { visited = append(visited, name); return false })
			assert.ElementsMatch(t, want, visited, sidecarDir)
		}
	})
	t.Run("Png", func(t *testing.T) {
		write(t, filepath.Join(sidecar, "2024", "logo.svg.png"), filepath.Join(sidecar, "2024", "logo.svg.PNG"), filepath.Join(sidecar, "2024", "logo.svg.apng"))
		var visited []string
		ImagePng.FindGenerated(filepath.Join(dir, "logo.svg"), dirs, originals, false, func(name string) bool { visited = append(visited, name); return false })
		assert.Equal(t, []string{filepath.Join(sidecar, "2024", "logo.svg.png")}, visited)
	})
	t.Run("EmptyBase", func(t *testing.T) {
		write(t, filepath.Join(originals, "2024.jpg"))
		assert.Empty(t, findAll(dir+string(os.PathSeparator)+"."))
	})
}

func TestFileType(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		result := FileType("")
		assert.Equal(t, TypeUnknown, result)
	})
	t.Run("Jpeg", func(t *testing.T) {
		result := FileType("testdata/test.jpg")
		assert.Equal(t, ImageJpeg, result)
	})
	t.Run("MpJpg", func(t *testing.T) {
		assert.Equal(t, ImageJpeg, FileType("name.MP.jpg"))
	})
	t.Run("RawCRW", func(t *testing.T) {
		result := FileType("testdata/test (jpg).crw")
		assert.Equal(t, ImageRaw, result)
	})
	t.Run("RawCR2", func(t *testing.T) {
		result := FileType("testdata/test (jpg).CR2")
		assert.Equal(t, ImageRaw, result)
	})
	t.Run("Mp4", func(t *testing.T) {
		assert.Equal(t, Type("mp4"), FileType("file.mp4"))
	})
	t.Run("Insp", func(t *testing.T) {
		assert.Equal(t, ImageInsp, FileType("IMG_20180318_205851_239.insp"))
	})
	t.Run("Insv", func(t *testing.T) {
		assert.Equal(t, VideoInsv, FileType("VID_20220607_102410_00_322.insv"))
	})
	t.Run("Lrv", func(t *testing.T) {
		assert.Equal(t, VideoLrv, FileType("LRV_20240415_213145_01_035.lrv"))
		assert.Equal(t, VideoLrv, FileType("LRV_20240415_213145_01_035.LRV"))
	})
	t.Run("Jpeg2000", func(t *testing.T) {
		// JPEG 2000 is not registered yet, and is never classified as ordinary JPEG. Every
		// extension of the family is listed, since one coder decodes them all and registering
		// any single one would reach it.
		for _, name := range []string{"scan.jp2", "scan.J2K", "scan.j2c", "scan.jpc", "scan.jpf", "scan.JPX", "scan.jpm"} {
			assert.Equalf(t, TypeUnknown, FileType(name), "%s must stay unregistered", name)
		}
	})
	t.Run("Cineon", func(t *testing.T) {
		assert.Equal(t, ImageCineon, FileType("frame.cin"))
		assert.Equal(t, ImageCineon, FileType("frame.CIN"))
	})
	t.Run("PhotoshopLargeDocument", func(t *testing.T) {
		assert.Equal(t, ImagePsd, FileType("artwork.psb"))
		assert.Equal(t, ImagePsd, FileType("artwork.PSB"))
	})
	t.Run("RawOlympusOri", func(t *testing.T) {
		assert.Equal(t, ImageRaw, FileType("P1010101.ori"))
		assert.Equal(t, ImageRaw, FileType("P1010101.ORI"))
	})
	t.Run("IllustratorTemplate", func(t *testing.T) {
		assert.Equal(t, VectorAI, FileType("logo.ait"))
		assert.Equal(t, VectorAI, FileType("logo.AIT"))
	})
	t.Run("MpegProgramStream", func(t *testing.T) {
		// DVD and SD camcorder recordings are MPEG program streams, like .mpg.
		assert.Equal(t, VideoMpeg, FileType("VTS_01_1.VOB"))
		assert.Equal(t, VideoMpeg, FileType("MOV001.mod"))
		assert.Equal(t, VideoMpeg, FileType("clip.mpe"))
	})
	t.Run("MpegVideoStream", func(t *testing.T) {
		assert.Equal(t, VideoMp2, FileType("clip.m2v"))
	})
	t.Run("MpegTransportStream", func(t *testing.T) {
		// JVC HD camcorder recordings are MPEG-2 transport streams, like .m2t.
		assert.Equal(t, VideoM2TS, FileType("MOV002.TOD"))
		assert.Equal(t, VideoM2TS, FileType("MOV002.tod"))
	})
	t.Run("DivX", func(t *testing.T) {
		assert.Equal(t, VideoAVI, FileType("movie.divx"))
		assert.Equal(t, VideoAVI, FileType("movie.DIVX"))
	})
	t.Run("Mobile", func(t *testing.T) {
		assert.Equal(t, Video3GP, FileType("clip.3gpp"))
		assert.Equal(t, Video3G2, FileType("clip.3gp2"))
	})
	t.Run("QuickTime", func(t *testing.T) {
		assert.Equal(t, VideoMov, FileType("clip.mqv"))
	})
	t.Run("MicroMV", func(t *testing.T) {
		// Sony MicroMV wraps MPEG-2 in a proprietary container that FFmpeg cannot demux.
		assert.Equal(t, TypeUnknown, FileType("clip.mmv"))
	})
}

func TestIsPreviewImageExt(t *testing.T) {
	for _, name := range []string{"photo.jpg", "photo.jpeg", "photo.JPEG", "photo.MP.jpg", "image.png", "image.PNG"} {
		assert.True(t, IsPreviewImageExt(name), name)
	}
	for _, name := range []string{"", "photo", "photo.webp", "photo.heic", "photo.dng", "photo.jpg.xmp", "video.mp4"} {
		assert.False(t, IsPreviewImageExt(name), name)
	}
}

func TestIsAnimatedImage(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, IsAnimatedImage(""))
	})
	t.Run("JPEG", func(t *testing.T) {
		assert.False(t, IsAnimatedImage("testdata/test.jpg"))
	})
	t.Run("RawCRW", func(t *testing.T) {
		assert.False(t, IsAnimatedImage("testdata/test (jpg).crw"))
	})
	t.Run("Mp4", func(t *testing.T) {
		assert.False(t, IsAnimatedImage("file.mp"))
		assert.False(t, IsAnimatedImage("file.mp4"))
	})
	t.Run("GIF", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.gif"))
	})
	t.Run("WebP", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.webp"))
	})
	t.Run("PNG", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.png"))
		assert.True(t, IsAnimatedImage("file.apng"))
		assert.True(t, IsAnimatedImage("file.pnga"))
	})
	t.Run("AVIF", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.avif"))
		assert.True(t, IsAnimatedImage("file.avis"))
		assert.True(t, IsAnimatedImage("file.avifs"))
	})
	t.Run("HEIC", func(t *testing.T) {
		assert.True(t, IsAnimatedImage("file.heic"))
		assert.True(t, IsAnimatedImage("file.heics"))
	})
}

// TestType_FindInSearchDir verifies that a file in a search folder is not looked up below that folder under its
// absolute path, while files outside the base folder keep their absolute fallback.
func TestType_FindInSearchDir(t *testing.T) {
	setup := func(t *testing.T) (originals, sidecar string) {
		t.Helper()
		root := t.TempDir()
		originals, sidecar = filepath.Join(root, "originals"), filepath.Join(root, "sidecar")
		require.NoError(t, os.MkdirAll(originals, 0o700))
		require.NoError(t, os.MkdirAll(sidecar, 0o700))
		return originals, sidecar
	}
	write := func(t *testing.T, fileName string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(filepath.Dir(fileName), 0o700))
		require.NoError(t, os.WriteFile(fileName, []byte("x"), 0o600))
	}

	t.Run("FileInSidecar", func(t *testing.T) {
		originals, sidecar := setup(t)
		fileName := filepath.Join(sidecar, "2024", "IMG_1.jpg")
		write(t, fileName)
		write(t, filepath.Join(sidecar, filepath.Dir(fileName), "IMG_1.json"))
		assert.Equal(t, "", SidecarJson.FindFirst(fileName, []string{sidecar}, originals, false))
		assert.Empty(t, SidecarJson.FindAll(fileName, []string{sidecar}, originals, false))
	})
	t.Run("FileOutsideBaseDir", func(t *testing.T) {
		originals, sidecar := setup(t)
		samples := filepath.Join(filepath.Dir(originals), "samples")
		fileName := filepath.Join(samples, "IMG_1.jpg")
		want := filepath.Join(sidecar, samples, "IMG_1.json")
		write(t, fileName)
		write(t, want)
		assert.Equal(t, want, SidecarJson.FindFirst(fileName, []string{sidecar}, originals, false))
	})
	t.Run("FileInOriginals", func(t *testing.T) {
		originals, sidecar := setup(t)
		fileName := filepath.Join(originals, "2024", "IMG_1.jpg")
		want := filepath.Join(sidecar, "2024", "IMG_1.json")
		write(t, fileName)
		write(t, want)
		assert.Equal(t, want, SidecarJson.FindFirst(fileName, []string{sidecar}, originals, false))
	})
	t.Run("SidecarInOriginals", func(t *testing.T) {
		originals, _ := setup(t)
		sidecar := filepath.Join(originals, ".photoprism", "storage", "sidecar")
		fileName := filepath.Join(sidecar, "2024", "IMG_1.jpg")
		want, err := FilePath(fileName, sidecar, originals, ".json")
		require.NoError(t, err)
		write(t, fileName)
		write(t, want)
		assert.Equal(t, want, SidecarJson.FindFirst(fileName, []string{sidecar}, originals, false))
	})
	t.Run("FilePathAgrees", func(t *testing.T) {
		originals, sidecar := setup(t)
		samples := filepath.Join(filepath.Dir(originals), "samples")

		for _, fileName := range []string{
			filepath.Join(sidecar, "2024", "IMG_1.jpg"),
			filepath.Join(samples, "IMG_1.jpg"),
			filepath.Join(originals, "2024", "IMG_1.jpg"),
			filepath.Join(originals, "IMG_1.jpg"),
		} {
			want, err := FilePath(fileName, sidecar, originals, ".json")
			require.NoError(t, err)
			write(t, fileName)
			write(t, want)
			assert.Equal(t, want, SidecarJson.FindFirst(fileName, []string{sidecar}, originals, false), fileName)
		}
	})
	t.Run("FileInOriginalsRoot", func(t *testing.T) {
		originals, sidecar := setup(t)
		fileName := filepath.Join(originals, "IMG_1.jpg")
		want := filepath.Join(sidecar, "IMG_1.json")
		write(t, fileName)
		write(t, want)
		assert.Equal(t, want, SidecarJson.FindFirst(fileName, []string{sidecar}, originals, false))
	})
}
