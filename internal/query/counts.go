package query

import "github.com/jinzhu/gorm"

// UpdatePhotoCounts updates photos count in related tables as needed.
func UpdatePhotoCounts() error {
	/*
	   UPDATE places
	   SET
	   photo_count = (SELECT
	   COUNT(*) FROM
	   photos ph
	   WHERE places.id = ph.place_id AND ph.photo_quality >= 0 AND ph.deleted_at IS NULL)
	*/

	return Db().Table("places").
		UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(*) FROM photos ph WHERE places.id = ph.place_id AND ph.photo_quality >= 0 AND ph.deleted_at IS NULL)")).Error
}
