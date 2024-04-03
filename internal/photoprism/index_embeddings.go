package photoprism

import (
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
)

func PopulateEmbeddingsIndex(conf *config.Config) error {
	log.Infof("embeddings: populating embeddings index")
	db := entity.Db()
	rows, err := db.Model(&entity.File{}).Where("photo_embeddings IS NOT NULL").Select("photo_id,photo_embeddings").Rows()
	defer rows.Close()
	if err != nil {
		return err
	}

	for rows.Next() {
		var file entity.File
		db.ScanRows(rows, &file)
		// FIXME: uint -> int64 ??
		id := int64(file.PhotoID)
		if len(file.PhotoEmbeddings) > 0 {
			conf.AddToEmbeddingsIndex(id, file.PhotoEmbeddings)
		}
	}
	return nil
}

// Embeddings caluclates embeddings vector for jpeg image
func (ind *Index) Embeddings(jpeg *MediaFile) (results []float32) {
	start := time.Now()

	filename, err := jpeg.Thumbnail(Config().ThumbCachePath(), thumb.Tile224)
	if err != nil {
		log.Debugf("%s in %s", err, clean.Log(jpeg.BaseName()))
	}

	results, err = ind.clipEmbeddings.File(filename)
	if err != nil {
		log.Debugf("%s in %s", err, clean.Log(jpeg.BaseName()))
	}

	log.Infof("embeddings: calculated embeddings vector %d for %s [%s]", len(results), clean.Log(jpeg.BaseName()), time.Since(start))
	return results
}
