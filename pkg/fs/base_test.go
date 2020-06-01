package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	t.Run("Screenshot 2019-05-21 at 10.45.52.png", func(t *testing.T) {
		regular := Base("Screenshot 2019-05-21 at 10.45.52.png", false)
		assert.Equal(t, "Screenshot 2019-05-21 at 10.45.52", regular)
		stripped := Base("Screenshot 2019-05-21 at 10.45.52.png", true)
		assert.Equal(t, "Screenshot 2019-05-21 at 10.45.52", stripped)
	})

	t.Run("Test.jpg", func(t *testing.T) {
		result := Base("/testdata/Test.jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := Base("/testdata/Test copy 3.jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := Base("/testdata/Test (3).jpg", true)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test.jpg", func(t *testing.T) {
		result := Base("/testdata/Test.jpg", false)
		assert.Equal(t, "Test", result)
	})

	t.Run("Test.3453453.jpg", func(t *testing.T) {
		regular := Base("/testdata/Test.3453453.jpg", false)
		assert.Equal(t, "Test.3453453", regular)

		stripped := Base("/testdata/Test.3453453.jpg", true)
		assert.Equal(t, "Test", stripped)
	})

	t.Run("/foo/bar.0000.ZIP", func(t *testing.T) {
		regular := Base("/foo/bar.0000.ZIP", false)
		assert.Equal(t, "bar.0000", regular)

		stripped := Base("/foo/bar.0000.ZIP", true)
		assert.Equal(t, "bar.0000", stripped)
	})

	t.Run("/foo/bar.00001.ZIP", func(t *testing.T) {
		regular := Base("/foo/bar.00001.ZIP", false)
		assert.Equal(t, "bar.00001", regular)

		stripped := Base("/foo/bar.00001.ZIP", true)
		assert.Equal(t, "bar", stripped)
	})

	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := Base("/testdata/Test copy 3.jpg", false)
		assert.Equal(t, "Test copy 3", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := Base("/testdata/Test (3).jpg", false)
		assert.Equal(t, "Test (3)", result)
	})
	t.Run("20180506_091537_DSC02122.JPG", func(t *testing.T) {
		result := Base("20180506_091537_DSC02122.JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("20180506_091537_DSC02122 (+3.3).JPG", func(t *testing.T) {
		result := Base("20180506_091537_DSC02122 (+3.3).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("20180506_091537_DSC02122 (-2.7).JPG", func(t *testing.T) {
		result := Base("20180506_091537_DSC02122 (-2.7).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("20180506_091537_DSC02122(+3.3).JPG", func(t *testing.T) {
		result := Base("20180506_091537_DSC02122(+3.3).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
	t.Run("20180506_091537_DSC02122(-2.7).JPG", func(t *testing.T) {
		result := Base("20180506_091537_DSC02122(-2.7).JPG", true)
		assert.Equal(t, "20180506_091537_DSC02122", result)
	})
}

func TestRelativeBase(t *testing.T) {
	t.Run("/foo/bar.0000.ZIP", func(t *testing.T) {
		regular := RelativeBase("/foo/bar.0000.ZIP", "/bar", false)
		assert.Equal(t, "/foo/bar.0000", regular)

		stripped := RelativeBase("/foo/bar.0000.ZIP", "/bar", true)
		assert.Equal(t, "/foo/bar.0000", stripped)
	})

	t.Run("/foo/bar.00001.ZIP", func(t *testing.T) {
		regular := RelativeBase("/foo/bar.00001.ZIP", "/bar", false)
		assert.Equal(t, "/foo/bar.00001", regular)

		stripped := RelativeBase("/foo/bar.00001.ZIP", "/bar", true)
		assert.Equal(t, "/foo/bar", stripped)
	})

	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := RelativeBase("/testdata/foo/Test copy 3.jpg", "/testdata", false)
		assert.Equal(t, "foo/Test copy 3", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := RelativeBase("/testdata/foo/Test (3).jpg", "/testdata", false)
		assert.Equal(t, "foo/Test (3)", result)
	})
}

func TestBaseAbs(t *testing.T) {
	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := AbsBase("/testdata/Test (4).jpg", true)

		assert.Equal(t, "/testdata/Test", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := AbsBase("/testdata/Test (4).jpg", false)

		assert.Equal(t, "/testdata/Test (4)", result)
	})

}

func TestSubFileName(t *testing.T) {
	t.Run("Test copy 3.jpg", func(t *testing.T) {
		result := SubFileName("/testdata/Test (4).jpg", ".photoprism", ".xmp", true)

		assert.Equal(t, "/testdata/.photoprism/Test.xmp", result)
	})

	t.Run("Test (3).jpg", func(t *testing.T) {
		result := SubFileName("/testdata/Test (4).jpg", ".photoprism", ".xmp", false)

		assert.Equal(t, "/testdata/.photoprism/Test (4).xmp", result)
	})

	t.Run("FOO.XMP", func(t *testing.T) {
		result := SubFileName("/testdata/FOO.XMP", ".photoprism", ".jpeg", true)

		assert.Equal(t, "/testdata/.photoprism/FOO.jpeg", result)
	})
}
