package query

import "github.com/jinzhu/gorm"

// UpdatePhotoCounts updates photos count in related tables as needed.
func UpdatePhotoCounts() error {
	if err := Db().Table("places").
		UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(*) FROM photos ph "+
			"WHERE places.id = ph.place_id "+
			"AND ph.photo_quality >= 0 "+
			"AND ph.photo_private = 0 "+
			"AND ph.deleted_at IS NULL)")).Error; err != nil {
		return err
	}

	if err := Db().Table("labels").
		UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(*) FROM photos_labels pl "+
			"JOIN photos ph ON pl.photo_id = ph.id "+
			"LEFT JOIN categories c ON c.label_id = pl.label_id "+
			"LEFT JOIN labels lc ON lc.id = c.category_id "+
			"WHERE pl.label_id = labels.id "+
			"AND lc.deleted_at IS NULL "+
			"AND pl.uncertainty < 100 "+
			"AND ph.photo_quality >= 0 "+
			"AND ph.photo_private = 0 "+
			"AND ph.deleted_at IS NULL)")).Error; err != nil {
		return err
	}

	return nil
}
