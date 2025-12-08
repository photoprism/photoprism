package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_Equal(t *testing.T) {
	t.Run("UnknownUnknown", func(t *testing.T) {
		assert.True(t, Unknown.Equal(""))
	})
	t.Run("ImageImage", func(t *testing.T) {
		assert.True(t, Image.Equal(Image.String()))
	})
	t.Run("VideoImage", func(t *testing.T) {
		assert.False(t, Video.Equal(Image.String()))
	})
	t.Run("SidecarUnknown", func(t *testing.T) {
		assert.False(t, Sidecar.Equal(Unknown.String()))
	})
}

func TestType_NotEqual(t *testing.T) {
	t.Run("UnknownUnknown", func(t *testing.T) {
		assert.False(t, Unknown.NotEqual(""))
	})
	t.Run("ImageImage", func(t *testing.T) {
		assert.False(t, Image.NotEqual(Image.String()))
	})
	t.Run("VideoImage", func(t *testing.T) {
		assert.True(t, Video.NotEqual(Image.String()))
	})
	t.Run("SidecarUnknown", func(t *testing.T) {
		assert.True(t, Sidecar.NotEqual(Unknown.String()))
	})
}

func TestType_IsMain(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		assert.False(t, Unknown.IsMain())
		assert.False(t, Unknown.IsArchive())
		assert.True(t, Unknown.IsSidecar())
		assert.True(t, Unknown.IsUnknown())
	})
	t.Run("Archive", func(t *testing.T) {
		assert.False(t, Archive.IsMain())
		assert.True(t, Archive.IsArchive())
		assert.False(t, Archive.IsSidecar())
		assert.False(t, Archive.IsUnknown())
	})
	t.Run("Sidecar", func(t *testing.T) {
		assert.False(t, Sidecar.IsMain())
		assert.False(t, Sidecar.IsArchive())
		assert.True(t, Sidecar.IsSidecar())
		assert.False(t, Sidecar.IsUnknown())
	})
	t.Run("Image", func(t *testing.T) {
		assert.True(t, Image.IsMain())
		assert.False(t, Image.IsArchive())
		assert.False(t, Image.IsSidecar())
		assert.False(t, Image.IsUnknown())
	})
	t.Run("Raw", func(t *testing.T) {
		assert.True(t, Raw.IsMain())
		assert.False(t, Raw.IsArchive())
		assert.False(t, Raw.IsSidecar())
		assert.False(t, Raw.IsUnknown())
	})
	t.Run("Live", func(t *testing.T) {
		assert.True(t, Live.IsMain())
		assert.False(t, Live.IsArchive())
		assert.False(t, Live.IsSidecar())
		assert.False(t, Live.IsUnknown())
	})
	t.Run("Video", func(t *testing.T) {
		assert.True(t, Video.IsMain())
		assert.False(t, Video.IsArchive())
		assert.False(t, Video.IsSidecar())
		assert.False(t, Video.IsUnknown())
	})
}

func TestType_Priority(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		assert.Equal(t, PriorityUnknown, Priority["foo"])
		assert.Equal(t, PriorityUnknown, Priority[Unknown])
		assert.Equal(t, PrioritySidecar, Priority[Sidecar])
		assert.Equal(t, PriorityArchive, Priority[Archive])
		assert.Equal(t, PriorityImage, Priority[Image])
		assert.Equal(t, PriorityMainMedia, Priority[Image])
		assert.Equal(t, Priority[Live], Priority[Live])
		assert.Equal(t, Priority[Animated], Priority[Animated])
		assert.Equal(t, Priority[Animated], Priority[Audio])
		assert.Equal(t, Priority[Animated], Priority[Document])
	})
	t.Run("Less", func(t *testing.T) {
		assert.Less(t, PriorityUnknown, Priority[Image])
		assert.Less(t, PriorityArchive, Priority[Image])
		assert.Less(t, PrioritySidecar, Priority[Image])
		assert.Less(t, PrioritySidecar, Priority[Raw])
		assert.Less(t, PriorityImage, Priority[Raw])
		assert.Less(t, PriorityImage, Priority[Live])
		assert.Less(t, PriorityImage, Priority[Video])
		assert.Less(t, PriorityMainMedia, Priority[Live])
		assert.Less(t, PriorityMainMedia, Priority[Video])
		assert.Less(t, Priority[Unknown], Priority[Image])
		assert.Less(t, Priority[Unknown], Priority[Live])
		assert.Less(t, Priority[Image], Priority[Live])
		assert.Less(t, Priority[Video], Priority[Live])
	})
}

func TestType_Unknown(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		assert.True(t, Unknown.IsUnknown())
		assert.True(t, Type("foo").IsUnknown())
	})
	t.Run("Image", func(t *testing.T) {
		assert.False(t, Image.IsUnknown())
		assert.True(t, Image.IsMain())
	})
	t.Run("Video", func(t *testing.T) {
		assert.False(t, Video.IsUnknown())
		assert.True(t, Video.IsMain())
	})
	t.Run("Sidecar", func(t *testing.T) {
		assert.False(t, Sidecar.IsUnknown())
		assert.False(t, Sidecar.IsMain())
	})
}
