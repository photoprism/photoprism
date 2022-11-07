package photoprism

import (
	"github.com/photoprism/photoprism/internal/thumb"
)

// Clips get clip embeddings from JPEG image and saves matching embedding to db.
func (ind *Index) Clips(jpeg *MediaFile, photoId uint) error {
	filename, err := jpeg.Thumbnail(ind.thumbPath(), thumb.Tile224)
	if err != nil {
		return err
	}
	return ind.clip.EncodeImageAndSave(filename, photoId)
}
