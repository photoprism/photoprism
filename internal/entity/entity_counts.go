package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/photoprism/photoprism/internal/mutex"
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
		AND ph.photo_quality > -1
		AND ph.photo_private = FALSE
		AND ph.deleted_at IS NULL GROUP BY l.id
		UNION ALL
		SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
		JOIN categories c ON c.category_id = l.id
		JOIN photos_labels pl ON pl.label_id = c.label_id
		JOIN photos ph ON pl.photo_id = ph.id
		WHERE pl.uncertainty < 100
		AND ph.photo_quality > -1
		AND ph.photo_private = FALSE
		AND ph.deleted_at IS NULL GROUP BY l.id) counts GROUP BY label_id
		`).Scan(&result).Error; err != nil {
		log.Errorf("label-count: %s", err.Error())
	}

	return result
}

// UpdatePlacesCounts updates the places photo counts.
func UpdatePlacesCounts() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	// Update places.
	res := Db().Table("places").
		Where("id is not null").
		UpdateColumn("photo_count", Db().Model(&Photos{}).Select("count(*)").
			Where("places.id = photos.place_id AND photos.photo_quality > -1 AND photos.photo_private = FALSE AND photos.deleted_at IS NULL"))

	if res.Error != nil {
		return res.Error
	}

	log.Debugf("counts: updated %s [%s]", english.Plural(int(res.RowsAffected), "place", "places"), time.Since(start))

	return nil
}

// UpdateSubjectCounts updates the subject file counts.
func UpdateSubjectCounts(public bool) (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()

	var res *gorm.DB

	subjTable := Subject{}.TableName()

	var photosJoin clause.Expr

	// Count people tagged on private pictures?
	// see https://github.com/photoprism/photoprism/issues/4238
	// and https://github.com/photoprism/photoprism/issues/2570#issuecomment-1231690056
	if public {
		photosJoin = gorm.Expr("p.id = f.photo_id AND p.deleted_at IS NULL AND p.photo_private = FALSE")
	} else {
		photosJoin = gorm.Expr("p.id = f.photo_id AND p.deleted_at IS NULL")
	}

	condition := gorm.Expr("subj_type = ?", SubjPerson)

	switch DbDialect() {
	case MySQL, Postgres:
		res = Db().Exec(`UPDATE ? LEFT JOIN (
		SELECT m.subj_uid, COUNT(DISTINCT f.id) AS subj_files, COUNT(DISTINCT f.photo_id) AS subj_photos
			FROM files f
			JOIN photos p ON ?			    
			JOIN markers m ON f.file_uid = m.file_uid AND m.subj_uid IS NOT NULL AND m.subj_uid <> '' AND m.subj_uid IS NOT NULL
			WHERE m.marker_invalid = FALSE AND f.deleted_at IS NULL GROUP BY m.subj_uid
		) b ON b.subj_uid = subjects.subj_uid
		SET subjects.file_count = CASE WHEN b.subj_files IS NULL THEN 0 ELSE b.subj_files END, 
			subjects.photo_count = CASE WHEN b.subj_photos IS NULL THEN 0 ELSE b.subj_photos END
		WHERE ?`, gorm.Expr(subjTable), photosJoin, condition)
	case SQLite3:
		// Update files count.
		res = Db().Table(subjTable).
			Where("subj_type = ?", SubjPerson).
			UpdateColumn("file_count", gorm.Expr("(SELECT COUNT(DISTINCT f.id)"+
				" FROM files f JOIN photos p ON ?"+
				" JOIN markers m ON f.file_uid = m.file_uid AND m.subj_uid = subjects.subj_uid"+
				" WHERE m.marker_invalid = FALSE AND f.deleted_at IS NULL)", photosJoin))

		// Update photo count.
		if res.Error != nil {
			return res.Error
		} else {
			photosRes := Db().Table(subjTable).
				Where("subj_type = ?", SubjPerson).
				UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(DISTINCT f.photo_id)"+
					" FROM files f JOIN photos p ON ?"+
					" JOIN markers m ON f.file_uid = m.file_uid AND m.subj_uid = subjects.subj_uid"+
					" WHERE m.marker_invalid = FALSE AND f.deleted_at IS NULL)", photosJoin))
			res.RowsAffected += photosRes.RowsAffected
		}
	default:
		return fmt.Errorf("sql: unsupported dialect %s", DbDialect())
	}

	if res.Error != nil {
		return res.Error
	}

	log.Debugf("counts: updated %s [%s]", english.Plural(int(res.RowsAffected), "subject", "subjects"), time.Since(start))

	return nil
}

// UpdateLabelCounts updates the label photo counts.
func UpdateLabelCounts() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	start := time.Now()
	var res *gorm.DB
	if IsDialect(MySQL) || IsDialect(Postgres) {
		res = Db().Exec(`UPDATE labels LEFT JOIN (
		SELECT p2.label_id, COUNT(DISTINCT photo_id) AS label_photos FROM (
			SELECT pl.label_id as label_id, p.id AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
			WHERE p.photo_quality > -1 AND p.photo_private = FALSE AND p.deleted_at IS NULL
			UNION
			SELECT c.category_id as label_id, p.id AS photo_id FROM photos p
				JOIN photos_labels pl ON pl.photo_id = p.id AND pl.uncertainty < 100
				JOIN categories c ON c.label_id = pl.label_id
			WHERE p.photo_quality > -1 AND p.photo_private = FALSE AND p.deleted_at IS NULL
			) p2 GROUP BY p2.label_id
		) b ON b.label_id = labels.id
		SET photo_count = CASE WHEN b.label_photos IS NULL THEN 0 ELSE b.label_photos END`)
	} else if IsDialect(SQLite3) {
		res = Db().
			Table("labels").
			Where("id is not null").
			UpdateColumn("photo_count",
				gorm.Expr(`(SELECT photo_count FROM (SELECT label_id, SUM(photo_count) AS photo_count FROM (
				SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
					JOIN photos_labels pl ON pl.label_id = l.id
					JOIN photos ph ON pl.photo_id = ph.id
					WHERE pl.uncertainty < 100
					AND ph.photo_quality > -1
					AND ph.photo_private = FALSE
					AND ph.deleted_at IS NULL GROUP BY l.id
					UNION ALL
					SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
					JOIN categories c ON c.category_id = l.id
					JOIN photos_labels pl ON pl.label_id = c.label_id
					JOIN photos ph ON pl.photo_id = ph.id
					WHERE pl.uncertainty < 100
					AND ph.photo_quality > -1
					AND ph.photo_private = FALSE
					AND ph.deleted_at IS NULL GROUP BY l.id) counts GROUP BY label_id) label_counts WHERE label_id = labels.id)`))
	} else {
		return fmt.Errorf("sql: unsupported dialect %s", DbDialect())
	}

	if res.Error != nil {
		return res.Error
	}

	log.Debugf("counts: updated %s [%s]", english.Plural(int(res.RowsAffected), "label", "labels"), time.Since(start))

	return nil
}

// UpdateCounts updates precalculated photo and file counts.
func UpdateCounts() (err error) {
	log.Debug("index: updating counts")

	if err = UpdatePlacesCounts(); err != nil {
		if strings.Contains(err.Error(), "Error 1054") {
			log.Errorf("counts: failed to update places, potentially incompatible database version")
			log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
			return nil
		}

		return fmt.Errorf("%s while updating places counts", err)
	}

	if err = UpdateSubjectCounts(true); err != nil {
		if strings.Contains(err.Error(), "Error 1054") {
			log.Errorf("counts: failed to update subjects, potentially incompatible database version")
			log.Errorf("%s see https://jira.mariadb.org/browse/MDEV-25362", err)
			return nil
		}

		return fmt.Errorf("%s while updating subject counts", err)
	}

	if err = UpdateLabelCounts(); err != nil {
		return fmt.Errorf("%s while updating label counts", err)
	}

	/* TODO: Slow with many photos due to missing index.
	start = time.Now()

	// Update calendar album visibility.
	switch DbDialect() {
	default:
		if err = UnscopedDb().Exec(`UPDATE albums SET deleted_at = ? WHERE album_type=? AND id NOT IN (
		SELECT a.id FROM albums a JOIN photos p ON a.album_month = MONTH(p.taken_at) AND a.album_year = YEAR(p.taken_at)
		AND p.deleted_at IS NULL AND p.photo_quality > -1 AND p.photo_private = FALSE WHERE album_type=? GROUP BY a.id)`,
			TimeStamp(), AlbumMonth, AlbumMonth).Error; err != nil {
			return err
		}
		if err = UnscopedDb().Exec(`UPDATE albums SET deleted_at = NULL WHERE album_type=? AND id IN (
		SELECT a.id FROM albums a JOIN photos p ON a.album_month = MONTH(p.taken_at) AND a.album_year = YEAR(p.taken_at)
		AND p.deleted_at IS NULL AND p.photo_quality > -1 AND p.photo_private = FALSE WHERE album_type=? GROUP BY a.id)`,
			AlbumMonth, AlbumMonth).Error; err != nil {
			return err
		}
	}

	log.Debugf("calendar: updating visibility completed [%s]", time.Since(start))
	*/

	return nil
}
