package query

import "github.com/photoprism/photoprism/internal/entity"

type Counts struct {
	Cameras        int `json:"cameras"`
	Lenses         int `json:"lenses"`
	Countries      int `json:"countries"`
	Photos         int `json:"photos"`
	Videos         int `json:"videos"`
	Hidden         int `json:"hidden"`
	Favorites      int `json:"favorites"`
	Private        int `json:"private"`
	Review         int `json:"review"`
	Stories        int `json:"stories"`
	Albums         int `json:"albums"`
	Folders        int `json:"folders"`
	Files          int `json:"files"`
	Moments        int `json:"moments"`
	Places         int `json:"places"`
	Labels         int `json:"labels"`
	LabelMaxPhotos int `json:"labelMaxPhotos"`
}

// Refresh updates the counts.
func (c *Counts) Refresh() {
	Db().Table("cameras").
		Where("camera_slug <> 'zz' AND camera_slug <> ''").
		Select("COUNT(*) AS cameras").
		Take(c)

	Db().Table("lenses").
		Where("lens_slug <> 'zz' AND lens_slug <> ''").
		Select("COUNT(*) AS lenses").
		Take(c)

	Db().Table("photos").
		Select("SUM(photo_type = 'video' AND photo_quality > -1 AND photo_private = 0) AS videos, " +
			"SUM(photo_quality > -1 AND photo_quality < 3 AND photo_private = 0) AS review, " +
			"SUM(photo_quality = -1) AS hidden, " +
			"SUM(photo_type NOT IN ('live', 'video') AND photo_quality > -1 AND photo_private = 0) AS photos, " +
			"SUM(photo_favorite = 1 AND photo_private = 0 AND photo_quality > -1) AS favorites, " +
			"SUM(photo_private = 1 AND photo_quality > -1) AS private").
		Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
		Where("deleted_at IS NULL").
		Take(c)

	Db().Table("labels").
		Select("MAX(photo_count) as label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 OR label_favorite = 1)").
		Take(c)

	Db().Table("albums").
		Select("SUM(album_type = ?) AS albums, SUM(album_type = ?) AS moments, "+
			"SUM(album_type = ?) AS folders",
			entity.AlbumManual, entity.AlbumMoment, entity.AlbumFolder).
		Where("deleted_at IS NULL").
		Take(c)

	Db().Table("files").
		Select("COUNT(*) AS files").
		Where("file_missing = 0 AND file_root = ?", entity.RootOriginals).
		Take(c)

	Db().Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(c)

	Db().Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id <> 'zz'").
		Take(c)
}
