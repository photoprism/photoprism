package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotosRequest_Empty(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		req := PhotosRequest{
			Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh0"},
			Values: &PhotosForm{},
		}

		assert.False(t, req.Empty())
	})
	t.Run("True", func(t *testing.T) {
		req := PhotosRequest{
			Photos: []string{},
			Values: &PhotosForm{},
		}

		assert.True(t, req.Empty())
	})
}

func TestPhotosRequest_Get(t *testing.T) {
	req := PhotosRequest{
		Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh0"},
		Values: &PhotosForm{},
	}

	resp := req.Get()
	assert.Equal(t, 2, len(resp))
}

func TestPhotosRequest_String(t *testing.T) {
	req := PhotosRequest{
		Photos: []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh0"},
		Values: &PhotosForm{},
	}

	resp := req.String()
	assert.Equal(t, "ps6sg6be2lvl0yh7, ps6sg6be2lvl0yh0", resp)
}
