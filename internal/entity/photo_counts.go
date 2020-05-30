package entity

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type LabelPhotoCount struct {
	LabelID    int
	PhotoCount int
}

type LabelPhotoCounts []LabelPhotoCount

func LabelCounts() LabelPhotoCounts {
	result := LabelPhotoCounts{}

	if err := UnscopedDb().Raw(`
		SELECT label_id, SUM(photo_count) AS photo_count FROM (
			SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
		JOIN photos_labels pl ON pl.label_id = l.id
		JOIN photos ph ON pl.photo_id = ph.id
		WHERE pl.uncertainty < 100
		AND ph.photo_quality >= 0
		AND ph.photo_private = 0
		AND ph.deleted_at IS NULL GROUP BY l.id
		UNION ALL
		SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
		JOIN categories c ON c.category_id = l.id
		JOIN photos_labels pl ON pl.label_id = c.label_id
		JOIN photos ph ON pl.photo_id = ph.id
		WHERE pl.uncertainty < 100
		AND ph.photo_quality >= 0
		AND ph.photo_private = 0
		AND ph.deleted_at IS NULL GROUP BY l.id) counts GROUP BY label_id
		`).Scan(&result).Error; err != nil {
		log.Errorf("label-count: %s", err.Error())
	}

	return result
}

// UpdatePhotoCounts updates photos count in related tables as needed.
func UpdatePhotoCounts() error {
	// log.Info("index: updating photo counts")

	if err := Db().Table("places").
		UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(*) FROM photos p "+
			"WHERE places.id = p.place_id "+
			"AND p.photo_quality >= 0 "+
			"AND p.photo_private = 0 "+
			"AND p.deleted_at IS NULL)")).Error; err != nil {
		return err
	}

	/* See internal/entity/views.go

	CREATE OR REPLACE VIEW label_counts AS
	SELECT label_id, SUM(photo_count) AS photo_count FROM (
	(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
	            JOIN photos_labels pl ON pl.label_id = l.id
	            JOIN photos ph ON pl.photo_id = ph.id
				WHERE pl.uncertainty < 100
				AND ph.photo_quality >= 0
				AND ph.photo_private = 0
				AND ph.deleted_at IS NULL GROUP BY l.id)
	UNION ALL
	(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
	            JOIN categories c ON c.category_id = l.id
	            JOIN photos_labels pl ON pl.label_id = c.label_id
	            JOIN photos ph ON pl.photo_id = ph.id
				WHERE pl.uncertainty < 100
				AND ph.photo_quality >= 0
				AND ph.photo_private = 0
				AND ph.deleted_at IS NULL GROUP BY l.id)) counts GROUP BY label_id
	*/

	/* TODO: Requires view support

	if err := Db().
		Table("labels").
		UpdateColumn("photo_count",
			gorm.Expr("(SELECT photo_count FROM label_counts WHERE label_id = labels.id)")).Error; err != nil {
		log.Warn(err)
	} */

	if IsDialect(MySQL) {
		if err := Db().
			Table("labels").
			UpdateColumn("photo_count",
				gorm.Expr(`(SELECT photo_count FROM (
			SELECT label_id, SUM(photo_count) AS photo_count FROM (
			(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
			            JOIN photos_labels pl ON pl.label_id = l.id
			            JOIN photos ph ON pl.photo_id = ph.id
						WHERE pl.uncertainty < 100
						AND ph.photo_quality >= 0
						AND ph.photo_private = 0
						AND ph.deleted_at IS NULL GROUP BY l.id)
			UNION ALL
			(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
			            JOIN categories c ON c.category_id = l.id
			            JOIN photos_labels pl ON pl.label_id = c.label_id
			            JOIN photos ph ON pl.photo_id = ph.id
						WHERE pl.uncertainty < 100
						AND ph.photo_quality >= 0
						AND ph.photo_private = 0
						AND ph.deleted_at IS NULL GROUP BY l.id)) counts GROUP BY label_id
			) label_counts WHERE label_id = labels.id)`)).Error; err != nil {
			return err
		}
	} else if IsDialect(SQLite) {
		if err := Db().
			Table("labels").
			UpdateColumn("photo_count",
				gorm.Expr(`(SELECT photo_count FROM (SELECT label_id, SUM(photo_count) AS photo_count FROM (
				SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
					JOIN photos_labels pl ON pl.label_id = l.id
					JOIN photos ph ON pl.photo_id = ph.id
					WHERE pl.uncertainty < 100
					AND ph.photo_quality >= 0
					AND ph.photo_private = 0
					AND ph.deleted_at IS NULL GROUP BY l.id
					UNION ALL
					SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
					JOIN categories c ON c.category_id = l.id
					JOIN photos_labels pl ON pl.label_id = c.label_id
					JOIN photos ph ON pl.photo_id = ph.id
					WHERE pl.uncertainty < 100
					AND ph.photo_quality >= 0
					AND ph.photo_private = 0
					AND ph.deleted_at IS NULL GROUP BY l.id) counts GROUP BY label_id) label_counts WHERE label_id = labels.id)`)).Error; err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown sql dialect %s", DbDialect())
	}

	return nil
}
