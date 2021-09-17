package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type LabelPhotoCount struct {
	LabelID    int
	PhotoCount int
}

type LabelPhotoCounts []LabelPhotoCount

// LabelCounts returns the number of photos for each label ID.
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

// UpdatePlacesPhotoCounts updates the places photo counts.
func UpdatePlacesPhotoCounts() (err error) {
	start := time.Now()

	// Update places.
	if err = Db().Table("places").
		UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(*) FROM photos p "+
			"WHERE places.id = p.place_id "+
			"AND p.photo_quality >= 0 "+
			"AND p.photo_private = 0 "+
			"AND p.deleted_at IS NULL)")).Error; err != nil {
		return err
	}

	log.Debugf("counts: updated places [%s]", time.Since(start))

	return nil
}

// UpdateSubjectFileCounts updates the subject file counts.
func UpdateSubjectFileCounts() (err error) {
	start := time.Now()

	// Update subjects.
	if err = Db().Table(Subject{}.TableName()).
		UpdateColumn("file_count", gorm.Expr("(SELECT COUNT(*) FROM files f "+
			fmt.Sprintf(
				"JOIN %s m ON f.file_uid = m.file_uid AND m.subj_uid = %s.subj_uid ",
				Marker{}.TableName(),
				Subject{}.TableName())+
			" WHERE m.marker_invalid = 0 AND f.deleted_at IS NULL)")).Error; err != nil {
		return err
	}

	log.Debugf("counts: updated subjects [%s]", time.Since(start))

	return nil
}

// UpdateLabelPhotoCounts updates the label photo counts.
func UpdateLabelPhotoCounts() (err error) {
	start := time.Now()

	if IsDialect(MySQL) {
		if err = Db().
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
		if err = Db().
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

	log.Debugf("counts: updated labels [%s]", time.Since(start))

	return nil
}

// UpdatePhotoCounts updates static photos counts and visibilities.
func UpdatePhotoCounts() (err error) {
	if err = UpdatePlacesPhotoCounts(); err != nil {
		return err
	}

	if err = UpdateSubjectFileCounts(); err != nil {
		return err
	}

	if err = UpdateLabelPhotoCounts(); err != nil {
		return err
	}

	/* TODO: Slow with many photos due to missing index.
	start = time.Now()

	// Update calendar album visibility.
	switch DbDialect() {
	default:
		if err = UnscopedDb().Exec(`UPDATE albums SET deleted_at = ? WHERE album_type=? AND id NOT IN (
		SELECT a.id FROM albums a JOIN photos p ON a.album_month = MONTH(p.taken_at) AND a.album_year = YEAR(p.taken_at)
		AND p.deleted_at IS NULL AND p.photo_quality > -1 AND p.photo_private = 0 WHERE album_type=? GROUP BY a.id)`,
			TimeStamp(), AlbumMonth, AlbumMonth).Error; err != nil {
			return err
		}
		if err = UnscopedDb().Exec(`UPDATE albums SET deleted_at = NULL WHERE album_type=? AND id IN (
		SELECT a.id FROM albums a JOIN photos p ON a.album_month = MONTH(p.taken_at) AND a.album_year = YEAR(p.taken_at)
		AND p.deleted_at IS NULL AND p.photo_quality > -1 AND p.photo_private = 0 WHERE album_type=? GROUP BY a.id)`,
			AlbumMonth, AlbumMonth).Error; err != nil {
			return err
		}
	}

	log.Debugf("calendar: updating visibility completed [%s]", time.Since(start))
	*/

	return nil
}
