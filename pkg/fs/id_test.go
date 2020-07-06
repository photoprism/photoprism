package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAsciiID(t *testing.T) {
	assert.False(t, IsAsciiID("lt9k3pw1wowuy3c2"))
	assert.False(t, IsAsciiID("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.False(t, IsAsciiID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	assert.False(t, IsAsciiID("55785BAC-9A4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsAsciiID("550e8400-e29b-11d4-a716-446655440000"))
	assert.False(t, IsAsciiID("IMG_0599.JPG"))
	assert.True(t, IsAsciiID("DSC10599"))
	assert.True(t, IsAsciiID("IQVG4929"))
	assert.False(t, IsAsciiID("DSC_0599"))
	assert.False(t, IsAsciiID("iqVG4929"))
	assert.False(t, IsAsciiID("20091117_203458_ERROR000"))
	assert.False(t, IsAsciiID("20091117_203458_12345678"))
	assert.True(t, IsAsciiID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.True(t, IsAsciiID("123"))
	assert.False(t, IsAsciiID("_"))
	assert.False(t, IsAsciiID(""))
	assert.False(t, IsAsciiID("20191117-153400-Central-Park-New-York-2019-3qy.mov"))
	assert.False(t, IsAsciiID("e98eb86480a72bd585d228a709f0622f90e86cbc.jpg"))
	assert.False(t, IsAsciiID("IMG_8115.jpg"))
	assert.False(t, IsAsciiID("01 Introduction Businessmodel.pdf"))
	assert.False(t, IsAsciiID("A regular file name with 121345678643 numbers"))
}

func TestIsID(t *testing.T) {
	assert.True(t, IsID("lt9k3pw1wowuy3c2"))
	assert.True(t, IsID("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.True(t, IsID("6ba7b810-9dad-11d1-80b4-00c04fd430c8"))
	assert.True(t, IsID("55785BAC-9A4B-4747-B090-EE123FFEE437"))
	assert.True(t, IsID("550e8400-e29b-11d4-a716-446655440000"))
	assert.True(t, IsID("IMG_0599.JPG"))
	assert.True(t, IsID("DSC10599"))
	assert.True(t, IsID("IQVG4929"))
	assert.True(t, IsID("20091117_203458_ERROR000"))
	assert.True(t, IsID("20091117_203458_12345678"))
	assert.True(t, IsID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.True(t, IsID("123"))
	assert.False(t, IsID("_"))
	assert.False(t, IsID(""))
	assert.False(t, IsID("20191117-153400-Central-Park-New-York-2019-3qy.mov"))
	assert.True(t, IsID("e98eb86480a72bd585d228a709f0622f90e86cbc.jpg"))
	assert.True(t, IsID("IMG_8115.jpg"))
	assert.False(t, IsID("01 Introduction Businessmodel.pdf"))
	assert.False(t, IsID("A regular file name with 121345678643 numbers"))
}
