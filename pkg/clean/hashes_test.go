package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskHashes(t *testing.T) {
	const (
		md5    = "9e107d9d372bb6826bd81d3542a419d6"
		sha1   = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"
		sha256 = "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592"
	)

	t.Run("SHA1", func(t *testing.T) {
		assert.Equal(t, "2ca***", MaskHashes(sha1))
	})
	t.Run("MD5", func(t *testing.T) {
		assert.Equal(t, "md5 9e1***", MaskHashes("md5 "+md5))
	})
	t.Run("SHA256", func(t *testing.T) {
		assert.Equal(t, "expected d7a*** got d7a***", MaskHashes("expected "+sha256+" got "+sha256))
	})
	t.Run("SHA512", func(t *testing.T) {
		assert.Equal(t, "(d7a***)", MaskHashes("("+sha256+sha256+")"))
	})
	t.Run("ThumbName", func(t *testing.T) {
		assert.Equal(t, "vips: failed to write /cache/2/c/a/2ca***_720x720_fit.jpg",
			MaskHashes("vips: failed to write /cache/2/c/a/"+sha1+"_720x720_fit.jpg"))
	})
	t.Run("UpperCase", func(t *testing.T) {
		assert.Equal(t, "thumb: 2CA*** not found", MaskHashes("thumb: "+strings.ToUpper(sha1)+" not found"))
	})
	t.Run("SeparatedRuns", func(t *testing.T) {
		assert.Equal(t, "2ca***_2ca***", MaskHashes(sha1+"_"+sha1))
	})
	t.Run("MultibyteNeighbors", func(t *testing.T) {
		assert.Equal(t, "ä2ca***ö › 9e1***", MaskHashes("ä"+sha1+"ö › "+md5))
	})
	t.Run("AdjacentHexLetters", func(t *testing.T) {
		assert.Equal(t, "cached fac***.jpg", MaskHashes("cached face"+sha1+".jpg"))
	})
	t.Run("ShorterRun", func(t *testing.T) {
		s := "photo " + md5[:31] + " and uid ps6sg6be2lvl0y12"
		assert.Equal(t, s, MaskHashes(s))
	})
	t.Run("UUID", func(t *testing.T) {
		s := "node 0196f2b4-8c3e-7a51-9d2f-5e4b3a2c1d0e"
		assert.Equal(t, s, MaskHashes(s))
	})
	t.Run("Unchanged", func(t *testing.T) {
		s := "index: added main jpg file 2015/11/reunion.jpg"
		assert.Equal(t, s, MaskHashes(s))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", MaskHashes(""))
	})
}

func TestIsHexByte(t *testing.T) {
	t.Run("Hex", func(t *testing.T) {
		for _, b := range []byte("0123456789abcdefABCDEF") {
			assert.True(t, isHexByte(b), string(b))
		}
	})
	t.Run("NotHex", func(t *testing.T) {
		for _, b := range []byte("gG_-/. zZ*") {
			assert.False(t, isHexByte(b), string(b))
		}
	})
}
