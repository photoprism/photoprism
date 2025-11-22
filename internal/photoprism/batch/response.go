package batch

import (
	"github.com/photoprism/photoprism/internal/entity/search"
)

// PhotosResponse represents the selected photo model data,
// and the values of the batch edit form.
type PhotosResponse struct {
	Models search.PhotoResults `json:"models"`
	Values *PhotosForm         `json:"values"`
}
