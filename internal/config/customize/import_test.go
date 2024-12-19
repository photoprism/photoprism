package customize

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestImportSettings(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		s := ImportSettings{}

		assert.IsType(t, ImportSettings{}, s)
		assert.Equal(t, "", s.Path)
		assert.Equal(t, RootPath, s.GetPath())
		assert.Equal(t, false, s.Move)
		assert.Equal(t, "", s.Dest)
		assert.Equal(t, DefaultImportDest, s.GetDest())

		pathName, fileName := s.GetDestName()

		assert.Equal(t, "2006/01", pathName)
		assert.Equal(t, "20060102_150405_", fileName)
	})
	t.Run("Customized", func(t *testing.T) {
		customPath := "/foo/bar"
		customDest := "2006/01/02/20060102_150405_82F63B78.jpg"

		s := ImportSettings{
			Path: customPath,
			Move: true,
			Dest: customDest,
		}

		assert.IsType(t, ImportSettings{}, s)
		assert.Equal(t, customPath, s.Path)
		assert.Equal(t, customPath, s.GetPath())
		assert.Equal(t, true, s.Move)
		assert.Equal(t, customDest, s.Dest)
		assert.Equal(t, customDest, s.GetDest())

		pathName, fileName := s.GetDestName()

		assert.Equal(t, "2006/01/02", pathName)
		assert.Equal(t, "20060102_150405_", fileName)
	})
	t.Run("InvalidDest", func(t *testing.T) {
		customPath := "/foo/bar"
		customDest := "/1/2/20060102_150405_82F63B78.jpg"

		s := ImportSettings{
			Path: customPath,
			Move: true,
			Dest: customDest,
		}

		assert.IsType(t, ImportSettings{}, s)
		assert.Equal(t, customPath, s.Path)
		assert.Equal(t, customPath, s.GetPath())
		assert.Equal(t, true, s.Move)
		assert.Equal(t, DefaultImportDest, s.GetDest())
		assert.Equal(t, "", s.Dest)
		assert.Equal(t, DefaultImportDest, s.GetDest())

		pathName, fileName := s.GetDestName()

		assert.Equal(t, "2006/01", pathName)
		assert.Equal(t, "20060102_150405_", fileName)
	})
}

func TestImportDestRegexp(t *testing.T) {
	// Must match "^\\D*\\d{2,14}[\\-/_].*\\d{2,14}.*(\\.ext|\\.EXT)$"
	t.Run("Valid", func(t *testing.T) {
		assert.True(t, ImportDestRegexp.MatchString("2006/01/20060102_150405_82f63b78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString(DefaultImportDest))
		t.Logf("%s -> %s", DefaultImportDest, time.Now().Format("2006/01_02_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString(time.Now().Format(DefaultImportDest)))
		assert.True(t, ImportDestRegexp.MatchString("foo/"+DefaultImportDest))
		assert.True(t, ImportDestRegexp.MatchString("/foo/"+DefaultImportDest))
		assert.True(t, ImportDestRegexp.MatchString("2006/01/20060102_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString("2006/01/02/150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString("foo/2006/01_02_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString("/foo/2006/01_02_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString("2006/2/20060102_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString("2006/2/20060102_150405_82f63b78.COUNT.ext"))
		assert.True(t, ImportDestRegexp.MatchString("2006/2_150405_82F63B78.jpg"))

		if m := ImportDestRegexp.FindStringSubmatch("2006/01/20060102_150405_82f63b78.00000.EXT"); m != nil {
			t.Logf("Matches: %#v", m)
			assert.Equal(t, "2006/01/20060102_150405_82f63b78.00000.EXT", m[0])
			assert.Equal(t, "2006/01/20060102_150405_", m[1])
			assert.Equal(t, "82f63b78", m[2])
			assert.Equal(t, ".00000", m[3])
			assert.Equal(t, ".EXT", m[4])
		}

		if m := ImportDestRegexp.FindStringSubmatch("2006/2/20060102_150405_82F63B78.jpg"); m != nil {
			t.Logf("Matches: %#v", m)
			assert.Equal(t, "2006/2/20060102_150405_82F63B78.jpg", m[0])
			assert.Equal(t, "2006/2/20060102_150405_", m[1])
			assert.Equal(t, "82F63B78", m[2])
			assert.Equal(t, "", m[3])
			assert.Equal(t, ".jpg", m[4])
		}

		if m := ImportDestRegexp.FindStringSubmatch("foo/2006/01_02_150405_82f63b78.COUNT.ext"); m != nil {
			t.Logf("Matches: %#v", m)
			assert.Equal(t, "foo/2006/01_02_150405_82f63b78.COUNT.ext", m[0])
			assert.Equal(t, "foo/2006/01_02_150405_", m[1])
			assert.Equal(t, "82f63b78", m[2])
			assert.Equal(t, ".COUNT", m[3])
			assert.Equal(t, ".ext", m[4])
		}

		assert.True(t, ImportDestRegexp.MatchString("2006/01/20060102_150405_82f63b78.00000.EXT"))
		t.Logf("2006/01/20060102_150405_82f63b78.00000.EXT -> %s", time.Now().Format("2006/01/20060102_150405_82f63b78.00000.EXT"))
		assert.True(t, ImportDestRegexp.MatchString(time.Now().Format("2006/01/20060102_150405_82f63b78.00000.EXT")))

		assert.True(t, ImportDestRegexp.MatchString("2006/01/02/20060102_150405_82F63B78.jpg"))
		t.Logf("2006/01/02/20060102_150405_82F63B78.jpg -> %s", time.Now().Format("2006/01/02/20060102_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString(time.Now().Format("2006/01/02/20060102_150405_82F63B78.jpg")))

		assert.True(t, ImportDestRegexp.MatchString("2006/01_02_150405_82F63B78.jpg"))
		t.Logf("2006/01_02_150405_82F63B78.jpg -> %s", time.Now().Format("2006/01_02_150405_82F63B78.jpg"))
		assert.True(t, ImportDestRegexp.MatchString(time.Now().Format("2006/01_02_150405_82F63B78.jpg")))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.False(t, ImportDestRegexp.MatchString(""))
		assert.False(t, ImportDestRegexp.MatchString(DefaultImportDest+"foobar"))
		assert.False(t, ImportDestRegexp.MatchString(DefaultImportDest+".foobar"))
		assert.False(t, ImportDestRegexp.MatchString("1/2/20060102_150405_82F63B78.jpg"))
		assert.False(t, ImportDestRegexp.MatchString("2006/2/CHECKSUM.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/2/bar.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/20060102_150405.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/20060102_150405_CRC32.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/20060102_150405.EXT"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/02/150405.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/02/150405.EXT"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01_02_150405.ext"))
		assert.False(t, ImportDestRegexp.MatchString("foo/2006/01_02_150405.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/02/150405-CRC32.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01_02_150405_CRC32.ext"))
		assert.False(t, ImportDestRegexp.MatchString("foo/2006/01_02_150405_CRC32.ext"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01/02/150405-CHECKSUM.EXT"))
		assert.False(t, ImportDestRegexp.MatchString("2006/01_02_150405 CHECKSUM.ext"))
	})
}
