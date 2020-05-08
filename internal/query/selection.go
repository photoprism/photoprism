package query

import (
	"errors"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// PhotoSelection returns all selected photos.
func PhotoSelection(f form.Selection) (results []entity.Photo, err error) {
	if f.Empty() {
		return results, errors.New("no photos selected")
	}

	s := Db().NewScope(nil).DB()

	s = s.Table("photos").
		Select("photos.*").
		Joins("LEFT JOIN photos_labels pl ON pl.photo_id = photos.id").
		Joins("LEFT JOIN labels l ON pl.label_id = l.id AND l.deleted_at IS NULL").
		Joins("LEFT JOIN categories c ON c.label_id = pl.label_id").
		Joins("LEFT JOIN labels lc ON lc.id = c.category_id AND lc.deleted_at IS NULL").
		Where("photos.deleted_at IS NULL").
		Group("photos.id")

	s = s.Where("photos.photo_uuid IN (?) OR l.label_uuid IN (?) OR lc.label_uuid IN (?)", f.Photos, f.Labels, f.Labels)

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
