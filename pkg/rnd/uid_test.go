package rnd

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsUnique(t *testing.T) {
	assert.True(t, IsUnique("ls6sg1e1wowuy3c2", 'l'))
	assert.True(t, IsUnique("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb", 'l'))
	assert.True(t, IsUnique("6ba7b810-9dad-11d1-80b4-00c04fd430c8", 'l'))
	assert.True(t, IsUnique("55785BAC-9A4B-4747-B090-EE123FFEE437", 'l'))
	assert.True(t, IsUnique("550e8400-e29b-11d4-a716-446655440000", 'l'))
	assert.False(t, IsUnique("4B1FEF2D1CF4A5BE38B263E0637EDEAD", 'l'))
	assert.False(t, IsUnique("123", '1'))
	assert.False(t, IsUnique("_", '_'))
	assert.False(t, IsUnique("", '_'))
}

func TestIsUID(t *testing.T) {
	prefix := byte('x')

	for n := 0; n < 10; n++ {
		s := GenerateUID(prefix)
		t.Logf("UID %d: %s", n, s)
		assert.True(t, IsUID(s, prefix))
	}

	assert.True(t, IsUID("ls6sg1e1wowuy3c2", 'l'))
	assert.False(t, IsUID("ls6sg1e1wowuy3c2123", 'l'))
	assert.False(t, IsUID("ls6sg1e1wowuy3c2123", 'l'))
	assert.False(t, IsUID("ls6sg1e1AAA-owuy3c2123", 'l'))
	assert.False(t, IsUID("", 'l'))
	assert.False(t, IsUID("ls6sg1e1w  ?owuy  3c2123", 'l'))
	assert.False(t, IsUID(RefID(""), 'r'))
}

func TestInvalidUID(t *testing.T) {
	prefix := byte('x')

	for n := 0; n < 10; n++ {
		id := GenerateUID(prefix)
		assert.False(t, InvalidUID(id, prefix))
	}

	assert.False(t, InvalidUID("ls6sg1e1wowuy3c2", 'l'))
	assert.True(t, InvalidUID("ls6sg1e1wowuy3c2123", 'l'))
	assert.True(t, InvalidUID("ls6sg1e1wowuy3c2123", 'l'))
	assert.True(t, InvalidUID("ls6sg1e1AAA-owuy3c2123", 'l'))
	assert.True(t, InvalidUID("", 'l'))
	assert.True(t, InvalidUID("ls6sg1e1w  ?owuy  3c2123", 'l'))
	assert.True(t, InvalidUID(RefID(""), 'r'))
}

func TestIsHex(t *testing.T) {
	assert.True(t, IsHex("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.True(t, IsHex("6ba7b810-9dad-11d1-80b4"))
	assert.False(t, IsHex("55785BAC-9A4B-4747-B090-GE123FFEE437"))
	assert.False(t, IsHex("550e8400-e29b-11d4-a716_446655440000"))
	assert.True(t, IsHex("4B1FEF2D1CF4A5BE38B263E0637EDEAD"))
	assert.False(t, IsHex(""))
}

func TestIsAlnum(t *testing.T) {
	assert.False(t, IsAlnum("dafbfeb8-a129-4e7c-9cf0-e7996a701cdb"))
	assert.True(t, IsAlnum("dafbe7996a701cdb"))
	assert.False(t, IsAlnum(""))
}

func TestGenerateUID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uid := GenerateUID('c')
		t.Logf("UID %d: %s", n, uid)
		assert.Equal(t, len(uid), 16)
		assert.True(t, IsUID(uid, 'c'))
	}
}

func BenchmarkGenerateUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		GenerateUID('x')
	}
}

func TestGenerateUID_Time(t *testing.T) {
	t.Run("2017", func(t *testing.T) {
		date := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		t.Logf("Unix Timestamp: %d", date.Unix())
		t.Logf("Date Encoded: %s", strconv.FormatInt(date.UTC().Unix(), 36))
		// coj2qo0...
		result := generateUID('c', date)
		t.Logf("2017: %s", result)

		if ts, err := strconv.ParseInt(result[1:7], 36, 64); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, date.Unix(), ts)
			t.Logf("UID Timestamp: %d", ts)
		}

		assert.Equal(t, "coj2qo0", result[:7])
	})
	t.Run("Now", func(t *testing.T) {
		date := time.Now().UTC()
		t.Logf("Unix Timestamp: %d", date.Unix())
		t.Logf("Date Encoded: %s", strconv.FormatInt(date.Unix(), 36))
		// coj2qo0...
		// cs6sfay
		result := generateUID('c', date)
		t.Logf("Now: %s", result)

		if ts, err := strconv.ParseInt(result[1:7], 36, 64); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, date.Unix(), ts)
			t.Logf("UID Timestamp: %d", ts)
		}

		assert.Equal(t, "c", result[:1])
	})
}
