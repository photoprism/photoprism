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
		UpdateColumn("photo_count", gorm.Expr(`(SELECT COUNT(DISTINCT ph.id) FROM labels l
			LEFT JOIN categories c ON c.category_id = l.id
            LEFT JOIN photos_labels pl ON pl.label_id = l.id
			LEFT JOIN photos_labels plc ON plc.label_id = c.label_id
            LEFT JOIN labels lc ON lc.id = plc.label_id
            LEFT JOIN photos ph ON pl.photo_id = ph.id OR plc.photo_id = ph.id
			WHERE l.id = labels.id 
			AND lc.deleted_at IS NULL
			AND (pl.uncertainty < 100 OR pl.uncertainty IS NULL)
			AND (plc.uncertainty < 100 OR plc.uncertainty IS NULL)
			AND ph.photo_quality >= 0
			AND ph.photo_private = 0
			AND ph.deleted_at IS NULL)`)).Error; err != nil {
		return err
	}

	return nil
}
