package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelValues(t *testing.T) {
	t.Run("NoInterface", func(t *testing.T) {
		m := Photo{}
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		assert.Error(t, err)
		assert.IsType(t, Values{}, values)
		assert.Len(t, keys, 0)
	})
	t.Run("NewPhoto", func(t *testing.T) {
		m := &Photo{}
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, Values{}, values)
	})
	t.Run("ExistingPhoto", func(t *testing.T) {
		m := PhotoFixtures.Pointer("Photo01")
		values, keys, err := ModelValues(m, "ID", "PhotoUID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 2)
		assert.NotNil(t, values)
		assert.IsType(t, Values{}, values)
	})
	t.Run("NewFace", func(t *testing.T) {
		m := &Face{}
		values, keys, err := ModelValues(m, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 0)
		assert.NotNil(t, values)
		assert.IsType(t, Values{}, values)
	})
	t.Run("ExistingFace", func(t *testing.T) {
		m := FaceFixtures.Pointer("john-doe")
		values, keys, err := ModelValues(m, "ID")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, keys, 1)
		assert.NotNil(t, values)
		assert.IsType(t, Values{}, values)
	})
	t.Run("ByteSlices", func(t *testing.T) {
		m := &Session{ID: "example", DataJSON: json.RawMessage(`{"groups":["example"]}`)}
		values, _, err := ModelValues(m, "ID")
		require.NoError(t, err)

		assert.Equal(t, json.RawMessage(`{"groups":["example"]}`), values["DataJSON"])
	})
	t.Run("RelationSlices", func(t *testing.T) {
		m := &Photo{Files: []File{{FileName: "example.jpg"}}}
		values, _, err := ModelValues(m, "ID")
		require.NoError(t, err)

		assert.NotContains(t, values, "Files")
	})
}

func TestReportValue(t *testing.T) {
	t.Run("RawMessage", func(t *testing.T) {
		assert.Equal(t, `{"a":1}`, reportValue(json.RawMessage(`{"a":1}`)))
	})
	t.Run("Bytes", func(t *testing.T) {
		assert.Equal(t, "abc", reportValue([]byte("abc")))
	})
	t.Run("Other", func(t *testing.T) {
		assert.Equal(t, `"abc"`, reportValue("abc"))
		assert.Equal(t, "5", reportValue(int64(5)))
	})
}
