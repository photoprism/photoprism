package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateAlbumDefaultCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumDefaultCovers())
}

func TestUpdateAlbumFolderCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumFolderCovers())
}

func TestUpdateAlbumMonthCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumMonthCovers())
}

func TestUpdateAlbumCovers(t *testing.T) {
	assert.NoError(t, UpdateAlbumCovers())
}

func TestUpdateLabelCovers(t *testing.T) {
	assert.NoError(t, UpdateLabelCovers())
}

func TestUpdateSubjectCovers(t *testing.T) {
	assert.NoError(t, UpdateSubjectCovers(false))
	assert.NoError(t, UpdateSubjectCovers(true))
}

func TestUpdateCovers(t *testing.T) {
	assert.NoError(t, UpdateCovers())
}
