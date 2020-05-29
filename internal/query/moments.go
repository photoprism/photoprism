package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/photoprism/photoprism/internal/maps"
)

// Moment contains photo counts per month and year
type Moment struct {
	PhotoCategory string `json:"Category"`
	PhotoCountry  string `json:"Country"`
	PhotoState    string `json:"State"`
	PhotoYear     int    `json:"Year"`
	PhotoMonth    int    `json:"Month"`
	PhotoCount    int    `json:"Count"`
}

var MomentCategory = map[string]string{
	"botanical garden": "Botanical Gardens",
	"nature reserve": "Nature Reserves",
	"bay": "Bays, Capes & Beaches",
	"beach": "Bays, Capes & Beaches",
	"cape": "Bays, Capes & Beaches",
}

// Slug returns an identifier string for a moment.
func (m Moment) Slug() string {
	return slug.Make(m.Title())
}

// Title returns an english title for the moment.
func (m Moment) Title() string {
	if m.PhotoYear == 0 && m.PhotoMonth == 0 {
		if m.PhotoCategory != "" {
			return MomentCategory[m.PhotoCategory]
		}

		country := maps.CountryName(m.PhotoCountry)

		if strings.Contains(m.PhotoState, country) {
			return m.PhotoState
		}

		if m.PhotoState == "" {
			return m.PhotoCountry
		}

		return fmt.Sprintf("%s / %s", m.PhotoState, country)
	}

	if m.PhotoCountry != "" && m.PhotoYear > 1900 && m.PhotoMonth == 0 {
		if m.PhotoState != "" {
			return fmt.Sprintf("%s / %s / %d", m.PhotoState, maps.CountryName(m.PhotoCountry), m.PhotoYear)
		}

		return fmt.Sprintf("%s %d", maps.CountryName(m.PhotoCountry), m.PhotoYear)
	}

	if m.PhotoYear > 1900 && m.PhotoMonth > 0 && m.PhotoMonth <= 12 {
		date := time.Date(m.PhotoYear, time.Month(m.PhotoMonth), 1, 0, 0, 0, 0, time.UTC)

		if m.PhotoCountry == "" {
			return date.Format("January 2006")
		}

		return fmt.Sprintf("%s / %s", maps.CountryName(m.PhotoCountry), date.Format("January 2006"))
	}

	if m.PhotoMonth > 0 && m.PhotoMonth <= 12 {
		return time.Month(m.PhotoMonth).String()
	}

	return maps.CountryName(m.PhotoCountry)
}

type Moments []Moment

// MomentsTime counts photos by month and year.
func MomentsTime(threshold int) (results Moments, err error) {
	db := UnscopedDb().Table("photos").
		Where("photos.photo_quality >= 3 AND deleted_at IS NULL AND photo_year > 0 AND photo_month > 0").
		Select("photos.photo_year, photos.photo_month, COUNT(*) AS photo_count").
		Group("photos.photo_year, photos.photo_month").
		Order("photos.photo_year DESC, photos.photo_month DESC").
		Having("photo_count >= ?", threshold)

	if err := db.Scan(&results).Error; err != nil {
		return results, err
	}

	return results, nil
}

// MomentsCountries returns the most popular countries by year.
func MomentsCountries(threshold int) (results Moments, err error) {
	db := UnscopedDb().Table("photos").
		Where("photos.photo_quality >= 3 AND deleted_at IS NULL AND photo_country <> 'zz' AND photo_year > 0").
		Select("photo_country, photo_year, COUNT(*) AS photo_count ").
		Group("photo_country, photo_year").
		Having("photo_count >= ?", threshold)

	if err := db.Scan(&results).Error; err != nil {
		return results, err
	}

	return results, nil
}

// MomentsStates returns the most popular states and countries by year.
func MomentsStates(threshold int) (results Moments, err error) {
	db := UnscopedDb().Table("photos").
		Joins("JOIN places p ON p.place_uid = photos.place_uid").
		Where("photos.photo_quality >= 3 AND photos.deleted_at IS NULL AND p.loc_state <> '' AND p.loc_country <> 'zz'").
		Select("p.loc_country AS photo_country, p.loc_state AS photo_state, COUNT(*) AS photo_count").
		Group("photo_country, photo_state").
		Having("photo_count >= ?", threshold)

	if err := db.Scan(&results).Error; err != nil {
		return results, err
	}

	return results, nil
}

// MomentsCategories returns the most popular photo categories.
func MomentsCategories(threshold int) (results Moments, err error) {
	var cats []string

	for cat, _ := range MomentCategory {
		cats = append(cats, cat)
	}

	db := UnscopedDb().Table("photos").
		Joins("JOIN locations l ON l.loc_uid = photos.loc_uid AND photos.loc_uid <> 'zz'").
		Where("photos.photo_quality >= 3 AND photos.deleted_at IS NULL AND l.loc_category IN (?)", cats).
		Select("l.loc_category AS photo_category, COUNT(*) AS photo_count").
		Group("photo_category").
		Having("photo_count >= ?", threshold)

	if err := db.Scan(&results).Error; err != nil {
		return results, err
	}

	return results, nil
}
