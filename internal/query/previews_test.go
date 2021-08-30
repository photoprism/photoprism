package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateAlbumDefaultPreviews(t *testing.T) {
	assert.NoError(t, UpdateAlbumDefaultPreviews())
}

func TestUpdateAlbumFolderPreviews(t *testing.T) {
	assert.NoError(t, UpdateAlbumFolderPreviews())
}

func TestUpdateAlbumMonthPreviews(t *testing.T) {
	assert.NoError(t, UpdateAlbumMonthPreviews())
}

func TestUpdateAlbumPreviews(t *testing.T) {
	assert.NoError(t, UpdateAlbumPreviews())
}

func TestUpdateLabelPreviews(t *testing.T) {
	assert.NoError(t, UpdateLabelPreviews())
}

func TestUpdateSubjectPreviews(t *testing.T) {
	assert.NoError(t, UpdateSubjectPreviews())
}

func TestUpdatePreviews(t *testing.T) {
	assert.NoError(t, UpdatePreviews())
}
