package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNearSQLCreator(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		if _, _, err := nearSQLCreator("", float64(0)); err != nil {
			assert.Equal(t, ErrNotFound, err)
		} else {
			t.Fail()
		}
	})

	t.Run("ps6sg6be2lvl0y24", func(t *testing.T) {
		if qs, values, err := nearSQLCreator("ps6sg6be2lvl0y24", float64(0)); err != nil {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, 2, len(values))
			assert.Equal(t, "photos.cell_id BETWEEN ? AND ?", qs)
			assert.Equal(t, "s2:85d1ea400004", values[0])
			assert.Equal(t, "s2:85d1ea800004", values[1])
		}
	})
	t.Run("ps6sg6byk7wrbk30", func(t *testing.T) {
		if qs, values, err := nearSQLCreator("ps6sg6byk7wrbk30", float64(0)); err != nil {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, 2, len(values))
			assert.Equal(t, "photos.cell_id BETWEEN ? AND ?", qs)
			assert.Equal(t, "s2:1ef75a400004", values[0])
			assert.Equal(t, "s2:1ef75a800004", values[1])
		}
	})
	t.Run("ps6sg6be2lvl0y24 pipe ps6sg6byk7wrbk30", func(t *testing.T) {
		if qs, values, err := nearSQLCreator("ps6sg6be2lvl0y24|ps6sg6byk7wrbk30", float64(0)); err != nil {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, 4, len(values))
			assert.Equal(t, "photos.cell_id BETWEEN ? AND ? OR photos.cell_id BETWEEN ? AND ?", qs)
			assert.Equal(t, "s2:85d1ea400004", values[0])
			assert.Equal(t, "s2:85d1ea800004", values[1])
			assert.Equal(t, "s2:1ef75a400004", values[2])
			assert.Equal(t, "s2:1ef75a800004", values[3])
		}
	})
	t.Run("ps6sg6be2lvl0y24 pipe ps6sg6byk7wrtest", func(t *testing.T) {
		if qs, values, err := nearSQLCreator("ps6sg6be2lvl0y24|ps6sg6byk7wrtest", float64(0)); err != nil {
			assert.Nil(t, err)
		} else {
			assert.Equal(t, 2, len(values))
			assert.Equal(t, "photos.cell_id BETWEEN ? AND ?", qs)
			assert.Equal(t, "s2:85d1ea400004", values[0])
			assert.Equal(t, "s2:85d1ea800004", values[1])
		}
	})
	t.Run("ps6sg6be2lvltest pipe ps6sg6byk7wrtest", func(t *testing.T) {
		if _, _, err := nearSQLCreator("ps6sg6be2lvltest|ps6sg6byk7wrtest", float64(0)); err != nil {
			assert.Equal(t, ErrNotFound, err)
		} else {
			t.Fail()
		}
	})
}
