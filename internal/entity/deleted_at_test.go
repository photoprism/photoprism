package entity

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm/schema"
)

func TestDeletedAtTime(t *testing.T) {
	now := time.Date(2022, 10, 18, 8, 15, 21, 0, time.UTC)
	n := DeletedAtTime(now)
	assert.True(t, n.Valid)
	assert.True(t, now.Equal(n.Time))
}

func TestDeletedAt_IsZero(t *testing.T) {
	assert.True(t, DeletedAt{}.IsZero())
	assert.False(t, DeletedAtTime(time.Now()).IsZero())
}

func TestDeletedAt_Ptr(t *testing.T) {
	assert.Nil(t, DeletedAt{}.Ptr())
	now := time.Now().UTC()
	p := DeletedAtTime(now).Ptr()
	require.NotNil(t, p)
	assert.True(t, now.Equal(*p))
}

func TestDeletedAt_Scan(t *testing.T) {
	var n DeletedAt
	require.NoError(t, n.Scan(nil))
	assert.False(t, n.Valid)
	now := time.Now().UTC()
	require.NoError(t, n.Scan(now))
	assert.True(t, n.Valid)
	assert.True(t, now.Equal(n.Time))
}

func TestDeletedAt_Value(t *testing.T) {
	v, err := DeletedAt{}.Value()
	require.NoError(t, err)
	assert.Nil(t, v)
	now := time.Now().UTC()
	v, err = DeletedAtTime(now).Value()
	require.NoError(t, err)
	assert.Equal(t, now, v)
}

func TestDeletedAt_MarshalJSON(t *testing.T) {
	type model struct {
		DeletedAt DeletedAt `json:"DeletedAt"`
		Omitted   DeletedAt `json:"Omitted,omitzero"`
	}

	t.Run("Null", func(t *testing.T) {
		b, err := json.Marshal(model{})
		require.NoError(t, err)
		assert.Equal(t, `{"DeletedAt":null}`, string(b))

		var m model
		require.NoError(t, json.Unmarshal([]byte(`{"DeletedAt":null}`), &m))
		assert.False(t, m.DeletedAt.Valid)
	})
	t.Run("Time", func(t *testing.T) {
		now := time.Date(2022, 10, 18, 8, 15, 21, 0, time.UTC)
		b, err := json.Marshal(model{DeletedAt: DeletedAtTime(now), Omitted: DeletedAtTime(now)})
		require.NoError(t, err)
		assert.Equal(t, `{"DeletedAt":"2022-10-18T08:15:21Z","Omitted":"2022-10-18T08:15:21Z"}`, string(b))

		var m model
		require.NoError(t, json.Unmarshal(b, &m))
		assert.True(t, m.DeletedAt.Valid)
		assert.True(t, now.Equal(m.DeletedAt.Time))
	})
	t.Run("Invalid", func(t *testing.T) {
		var m model
		assert.Error(t, json.Unmarshal([]byte(`{"DeletedAt":"yesterday"}`), &m))
	})
}

func TestDeletedAt_MarshalYAML(t *testing.T) {
	type model struct {
		DeletedAt DeletedAt `yaml:"DeletedAt,omitempty"`
	}

	t.Run("Omitted", func(t *testing.T) {
		b, err := yaml.Marshal(model{})
		require.NoError(t, err)
		assert.Equal(t, "{}\n", string(b))

		var m model
		require.NoError(t, yaml.Unmarshal([]byte("DeletedAt: null\n"), &m))
		assert.False(t, m.DeletedAt.Valid)
	})
	t.Run("Time", func(t *testing.T) {
		now := time.Date(2022, 10, 18, 8, 15, 21, 0, time.UTC)
		b, err := yaml.Marshal(model{DeletedAt: DeletedAtTime(now)})
		require.NoError(t, err)
		assert.Equal(t, "DeletedAt: 2022-10-18T08:15:21Z\n", string(b))

		var m model
		require.NoError(t, yaml.Unmarshal(b, &m))
		assert.True(t, m.DeletedAt.Valid)
		assert.True(t, now.Equal(m.DeletedAt.Time))
	})
	t.Run("Invalid", func(t *testing.T) {
		var m model
		assert.Error(t, yaml.Unmarshal([]byte("DeletedAt: [1, 2]\n"), &m))
	})
}

func TestDeletedAt_Clauses(t *testing.T) {
	s, err := schema.Parse(&Photo{}, &sync.Map{}, Db().NamingStrategy)
	require.NoError(t, err)
	field := s.LookUpField("DeletedAt")
	require.NotNil(t, field)

	assert.NotEmpty(t, DeletedAt{}.QueryClauses(field))
	assert.NotEmpty(t, DeletedAt{}.UpdateClauses(field))
	assert.NotEmpty(t, DeletedAt{}.DeleteClauses(field))
}

func TestDeletedAt_SoftDelete(t *testing.T) {
	label := NewLabel("Soft Delete "+time.Now().UTC().Format("150405.000000"), 0)
	require.NoError(t, label.Save())
	t.Cleanup(func() { _ = UnscopedDb().Delete(label).Error })

	require.NoError(t, Db().Delete(label).Error)

	var count int64
	require.NoError(t, Db().Model(&Label{}).Where("id = ?", label.ID).Count(&count).Error)
	assert.Zero(t, count, "scoped query excludes the row")

	found := Label{}
	require.NoError(t, UnscopedDb().Where("id = ?", label.ID).First(&found).Error)
	assert.True(t, found.DeletedAt.Valid)
}
