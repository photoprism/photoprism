package entity

import (
	"errors"
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
func (m *Photo) Optimize(stackMeta, stackUuid bool) (updated bool, merged Photos, err error) {
	if !m.HasID() {
		return false, merged, errors.New("photo: can't maintain, id is empty")
	}

	current := *m

	if m.HasLatLng() && !m.HasLocation() {
		m.UpdateLocation()
	}

	if merged, err = m.Stack(stackMeta, stackUuid); err != nil {
		log.Errorf("photo: %s (stack)", err)
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

	if err := Db().First(&file, "file_primary = 1 AND photo_id = ?", m.ID).Error; err == nil {
		return file.ResolvePrimary()
	}

	if err := Db().First(&file, "file_type = 'jpg' AND photo_id = ?", m.ID).Error; err == nil {
		file.FilePrimary = true
		return file.ResolvePrimary()
	}

	return m.Update("PhotoQuality", -1)
}

// Stack merges a photo with identical ones.
func (m *Photo) Stack(stackMeta, stackUuid bool) (identical Photos, err error) {
	if !stackMeta && !stackUuid || m.PhotoSingle || m.DeletedAt != nil {
		return identical, nil
	}

	switch {
	case stackMeta && stackUuid && m.HasLocation() && m.HasLatLng() && m.TakenSrc == SrcMeta && rnd.IsUUID(m.UUID):
		if err := Db().Where("id > ? AND photo_single = 0", m.ID).
			Where("(taken_at = ? AND taken_src = 'meta' AND cell_id = ? AND camera_serial = ? AND camera_id = ?) OR (uuid <> '' AND uuid = ?)",
				m.TakenAt, m.CellID, m.CameraSerial, m.CameraID, m.UUID).Find(&identical).Error; err != nil {
			return identical, err
		}
	case stackMeta && m.HasLocation() && m.HasLatLng() && m.TakenSrc == SrcMeta:
		if err := Db().Where("id > ? AND photo_single = 0", m.ID).
			Where("taken_at = ? AND taken_src = 'meta' AND cell_id = ? AND camera_serial = ? AND camera_id = ?",
				m.TakenAt, m.CellID, m.CameraSerial, m.CameraID).Error; err != nil {
			return identical, err
		}
	case stackUuid && rnd.IsUUID(m.UUID):
		if err := Db().Where("id > ? AND photo_single = 0", m.ID).
			Where("uuid <> '' AND uuid = ?", m.UUID).Error; err != nil {
			return identical, err
		}
	default:
		return identical, nil
	}

	if len(identical) == 0 {
		return identical, nil
	}

	for _, photo := range identical {
		if err := UnscopedDb().Exec("UPDATE `files` SET photo_id = ?, photo_uid = ?, file_primary = 0 WHERE photo_id = ?", m.ID, m.PhotoUID, photo.ID).Error; err != nil {
			return identical, err
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
			log.Warnf("photo: unknown SQL dialect (stack)")
		}

		if err := photo.Updates(map[string]interface{}{"DeletedAt": Timestamp(), "PhotoQuality": -1}); err != nil {
			return identical, err
		}
	}

	return identical, err
}
