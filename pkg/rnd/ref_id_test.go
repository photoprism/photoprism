package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefID(t *testing.T) {
	id1 := RefID("")
	t.Logf("RefID 1: %s", id1)
	assert.NotEmpty(t, id1)
	assert.True(t, IsRefID(id1))
	assert.False(t, InvalidRefID(id1))

	id2 := RefID("us")
	t.Logf("RefID 2: %s", id2)
	assert.NotEmpty(t, id2)
	assert.True(t, IsRefID(id2))
	assert.False(t, InvalidRefID(id2))
	assert.Len(t, id2, 12)

	id3 := RefID("f")
	t.Logf("RefID 3: %s", id3)
	assert.NotEmpty(t, id3)
	assert.True(t, IsRefID(id3))
	assert.False(t, InvalidRefID(id3))
	assert.Len(t, id3, 12)

	id4 := RefID("sess")
	t.Logf("RefID 4: %s", id4)
	assert.NotEmpty(t, id4)
	assert.True(t, IsRefID(id4))
	assert.False(t, InvalidRefID(id4))
	assert.Len(t, id3, 12)

	id5 := RefID("ph")
	t.Logf("RefID 5: %s", id5)
	assert.NotEmpty(t, id5)
	assert.True(t, IsRefID(id5))
	assert.False(t, InvalidRefID(id5))
	assert.Len(t, id5, 12)

	id6 := RefID("alb")
	t.Logf("RefID 6: %s", id6)
	assert.NotEmpty(t, id6)
	assert.True(t, IsRefID(id6))
	assert.False(t, InvalidRefID(id6))
	assert.Len(t, id6, 12)

	id7 := RefID("abcdef")
	t.Logf("RefID 7: %s", id7)
	assert.NotEmpty(t, id7)
	assert.True(t, IsRefID(id7))
	assert.False(t, InvalidRefID(id7))
	assert.Len(t, id7, 12)

	for n := 7; n < 20; n++ {
		id := RefID("test")
		t.Logf("RefID %d: %s", n, id)
		assert.NotEmpty(t, id)
	}
}

func TestIsRefID(t *testing.T) {
	assert.True(t, IsRefID("azn4qiw843"))
	assert.True(t, IsRefID("abzn4qiw843"))
	assert.True(t, IsRefID("abczn4qiw843"))
	assert.True(t, IsRefID(RefID("")))
	assert.True(t, IsRefID(RefID("")))
	assert.False(t, IsRefID("gzn4q w8"))
	assert.False(t, IsRefID("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.False(t, IsRefID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
}

func TestInvalidRefID(t *testing.T) {
	assert.False(t, InvalidRefID("gzn4qiw843"))
	assert.False(t, InvalidRefID(RefID("")))
	assert.False(t, InvalidRefID(RefID("")))
	assert.True(t, InvalidRefID("gzn4q w8"))
	assert.True(t, InvalidRefID("55785BAC-9H4B-4747-B090-EE123FFEE437"))
	assert.True(t, InvalidRefID("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
}
