package entity

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// EstimateCountry updates the photo with an estimated country if possible.
func (m *Photo) EstimateCountry() {
	if m.HasLatLng() || m.HasLocation() || m.HasPlace() || m.HasCountry() && m.PlaceSrc != SrcAuto && m.PlaceSrc != SrcEstimate {
		// Do nothing.
		return
	}

	unknown := UnknownCountry.ID
	countryCode := unknown

	if code := txt.CountryCode(m.PhotoTitle); code != unknown {
		countryCode = code
	}

	if countryCode == unknown {
		if code := txt.CountryCode(m.PhotoName); code != unknown && !fs.IsGenerated(m.PhotoName) {
			countryCode = code
		} else if code := txt.CountryCode(m.PhotoPath); code != unknown {
			countryCode = code
		}
	}

	if countryCode == unknown && m.OriginalName != "" && !fs.IsGenerated(m.OriginalName) {
		if code := txt.CountryCode(m.OriginalName); code != UnknownCountry.ID {
			countryCode = code
		}
	}

	if countryCode != unknown {
		m.PhotoCountry = countryCode
		m.PlaceSrc = SrcEstimate
		log.Debugf("photo: probable country for %s is %s", m, txt.Quote(m.CountryName()))
	}
}

// EstimatePlace updates the photo with an estimated place and country if possible.
func (m *Photo) EstimatePlace() {
	if m.HasLatLng() || m.HasLocation() || m.HasPlace() && m.PlaceSrc != SrcAuto && m.PlaceSrc != SrcEstimate {
		// Do nothing.
		return
	}

	var recentPhoto Photo
	var dateExpr string

	switch DbDialect() {
	case MySQL:
		dateExpr = "ABS(DATEDIFF(taken_at, ?)) ASC"
	case SQLite:
		dateExpr = "ABS(JulianDay(taken_at) - JulianDay(?)) ASC"
	default:
		log.Errorf("photo: unknown sql dialect %s", DbDialect())
		return
	}

	if err := UnscopedDb().
		Where("place_id <> '' AND place_id <> 'zz' AND place_src <> '' AND place_src <> ?", SrcEstimate).
		Order(gorm.Expr(dateExpr, m.TakenAt)).
		Preload("Place").First(&recentPhoto).Error; err != nil {
		log.Debugf("photo: can't estimate place at %s", m.TakenAt)
		m.EstimateCountry()
	} else {
		if hours := recentPhoto.TakenAt.Sub(m.TakenAt) / time.Hour; hours < -36 || hours > 36 {
			log.Debugf("photo: can't estimate position of %s, %d hours time difference", m, hours)
		} else if recentPhoto.HasPlace() {
			m.Place = recentPhoto.Place
			m.PlaceID = recentPhoto.PlaceID
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.PlaceSrc = SrcEstimate
			log.Debugf("photo: approximate position of %s is %s (id %s)", m, txt.Quote(m.CountryName()), recentPhoto.PlaceID)
		} else if recentPhoto.HasCountry() {
			m.PhotoCountry = recentPhoto.PhotoCountry
			m.PlaceSrc = SrcEstimate
			log.Debugf("photo: probable country for %s is %s", m, txt.Quote(m.CountryName()))
		} else {
			m.EstimateCountry()
		}
	}
}

// Optimize photo data, improve if possible.
func (m *Photo) Optimize(mergeMeta, mergeUuid bool) (updated bool, merged Photos, err error) {
	if !m.HasID() {
		return false, merged, errors.New("photo: can't maintain, id is empty")
	}

	current := *m

	if m.HasLatLng() && !m.HasLocation() {
		m.UpdateLocation()
	}

	if merged, err = m.Merge(mergeMeta, mergeUuid, true); err != nil {
		log.Errorf("photo: %s (merge)", err)
	}

	m.EstimatePlace()

	labels := m.ClassifyLabels()

	m.UpdateDateFields()

	if err := m.UpdateTitle(labels); err != nil {
		log.Info(err)
	}

	details := m.GetDetails()
	w := txt.UniqueWords(txt.Words(details.Keywords))
	w = append(w, labels.Keywords()...)
	details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	m.PhotoQuality = m.QualityScore()

	checked := Timestamp()

	if reflect.DeepEqual(*m, current) {
		return false, merged, m.Update("CheckedAt", &checked)
	}

	m.CheckedAt = &checked

	return true, merged, m.Save()
}

// ResolvePrimary ensures there is only one primary file for a photo.
func (m *Photo) ResolvePrimary() error {
	var file File

	if err := Db().Where("file_primary = 1 AND photo_id = ?", m.ID).First(&file).Error; err == nil && file.ID > 0 {
		return file.ResolvePrimary()
	}

	return nil
}

// Identical returns identical photos that can be merged.
func (m *Photo) Identical(findMeta, findUuid, findOlder bool) (identical Photos, err error) {
	if !findMeta && !findUuid || m.PhotoSingle || m.DeletedAt != nil {
		return identical, nil
	}

	op := "<>"

	if findOlder {
		op = "<"
	}

	switch {
	case findMeta && findUuid && m.HasLocation() && m.HasLatLng() && m.TakenSrc == SrcMeta && rnd.IsUUID(m.UUID):
		if err := Db().
			Where("(taken_at = ? AND taken_src = 'meta' AND cell_id = ? AND camera_serial = ? AND camera_id = ?) OR (uuid <> '' AND uuid = ?)",
				m.TakenAt, m.CellID, m.CameraSerial, m.CameraID, m.UUID).
			Where(fmt.Sprintf("id %s ? AND photo_single = 0 AND deleted_at IS NULL AND edited_at IS NULL", op), m.ID).
			Order("id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	case findMeta && m.HasLocation() && m.HasLatLng() && m.TakenSrc == SrcMeta:
		if err := Db().
			Where("taken_at = ? AND taken_src = 'meta' AND cell_id = ? AND camera_serial = ? AND camera_id = ?",
				m.TakenAt, m.CellID, m.CameraSerial, m.CameraID).
			Where(fmt.Sprintf("id %s ? AND photo_single = 0 AND deleted_at IS NULL AND edited_at IS NULL", op), m.ID).
			Order("id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	case findUuid && rnd.IsUUID(m.UUID):
		if err := Db().
			Where(fmt.Sprintf("uuid = ? AND id %s ? AND photo_single = 0 AND deleted_at IS NULL AND edited_at IS NULL", op), m.UUID, m.ID).
			Order("id ASC").Find(&identical).Error; err != nil {
			return identical, err
		}
	}

	return identical, nil
}

// Merge photo with identical ones.
func (m *Photo) Merge(mergeMeta, mergeUuid, mergeOlder bool) (merged Photos, err error) {
	merged, err = m.Identical(mergeMeta, mergeUuid, mergeOlder)

	if len(merged) == 0 || err != nil {
		return merged, err
	}

	for _, photo := range merged {
		if photo.DeletedAt != nil || photo.ID == m.ID {
			continue
		}

		if err := UnscopedDb().Exec("UPDATE `files` SET photo_id = ?, photo_uid = ?, file_primary = 0 WHERE photo_id = ?", m.ID, m.PhotoUID, photo.ID).Error; err != nil {
			return merged, err
		}

		switch DbDialect() {
		case MySQL:
			UnscopedDb().Exec("UPDATE IGNORE `photos_keywords` SET `photo_id` = ? WHERE (photo_id = ?)", m.ID, photo.ID)
			UnscopedDb().Exec("UPDATE IGNORE `photos_labels` SET `photo_id` = ? WHERE (photo_id = ?)", m.ID, photo.ID)
			UnscopedDb().Exec("UPDATE IGNORE `photos_albums` SET `photo_uid` = ? WHERE (photo_uid = ?)", m.PhotoUID, photo.PhotoUID)
		case SQLite:
			UnscopedDb().Exec("UPDATE OR IGNORE `photos_keywords` SET `photo_id` = ? WHERE (photo_id = ?)", m.ID, photo.ID)
			UnscopedDb().Exec("UPDATE OR IGNORE `photos_labels` SET `photo_id` = ? WHERE (photo_id = ?)", m.ID, photo.ID)
			UnscopedDb().Exec("UPDATE OR IGNORE `photos_albums` SET `photo_uid` = ? WHERE (photo_uid = ?)", m.PhotoUID, photo.PhotoUID)
		default:
			log.Warnf("photo: unknown SQL dialect (merge)")
		}

		deleted := Timestamp()

		if err := UnscopedDb().Exec("UPDATE `photos` SET photo_quality = -1, deleted_at = ? WHERE id = ?", Timestamp(), photo.ID).Error; err != nil {
			return merged, err
		}

		photo.DeletedAt = &deleted
		photo.PhotoQuality = -1
	}

	return merged, err
}
