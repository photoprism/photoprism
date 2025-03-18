package entity

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// HasTitle checks if the photo has a title.
func (m *Photo) HasTitle() bool {
	return m.PhotoTitle != ""
}

// NoTitle checks if the photo has no Title
func (m *Photo) NoTitle() bool {
	return m.PhotoTitle == ""
}

// GetTitle returns the photo title, if any.
func (m *Photo) GetTitle() string {
	return m.PhotoTitle
}

// SetTitle changes the photo title and clips it to 300 characters.
func (m *Photo) SetTitle(title, source string) {
	title = strings.Trim(title, "_&|{}<>: \n\r\t\\")
	title = strings.ReplaceAll(title, "\"", "'")
	title = txt.Shorten(title, txt.ClipLongName, txt.Ellipsis)

	// Get source priority.
	p := SrcPriority[source]

	// Compare the source priority with the priority of the current title source.
	if (p < SrcPriority[m.TitleSrc]) && m.HasTitle() {
		return
	}

	// Allow users to manually delete existing titles.
	if title == "" && p != 1 && p < SrcPriority[SrcManual] {
		return
	}

	m.PhotoTitle = title
	m.TitleSrc = source
}

// GenerateTitle tries to generate a title based on the picture
// location and labels if no other title is currently set.
func (m *Photo) GenerateTitle(labels classify.Labels) error {
	if m.TitleSrc != SrcAuto {
		return fmt.Errorf("photo: %s keeps existing %s title", m.String(), SrcString(m.TitleSrc))
	}

	var names string
	var knownLocation bool

	start := time.Now()
	oldTitle := m.PhotoTitle
	fileTitle := m.FileTitle()

	// Find people in the picture.
	people := m.SubjectNames()

	m.GenerateCaption(people)

	if n := len(people); n > 0 && n < 4 {
		names = txt.JoinNames(people, false)
	}

	if m.LocationLoaded() && m.TrustedLocation() {
		knownLocation = true
		loc := m.Cell

		// TODO: User defined format for generated titles.
		if names != "" {
			log.Debugf("photo: %s title based on %s (%s)", m.String(), english.Plural(len(people), "person", "people"), clean.Log(names))

			if l := len([]rune(names)); l > 35 {
				m.SetTitle(names, SrcAuto)
			} else if l > 20 && (loc.NoCity() || loc.LongCity()) {
				m.SetTitle(fmt.Sprintf("%s / %s", names, m.TakenAt.Format("2006")), SrcAuto)
			} else if l > 20 {
				m.SetTitle(fmt.Sprintf("%s / %s", names, loc.City()), SrcAuto)
			} else if loc.NoCity() || loc.LongCity() {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", names, loc.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", names, loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if title := labels.Title(loc.Name()); title != "" {
			log.Debugf("photo: %s title based on label %s", m.String(), clean.Log(title))
			if loc.NoCity() || loc.LongCity() || loc.CityContains(title) {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), loc.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if loc.Name() != "" && loc.City() != "" {
			if len(loc.Name()) > 45 {
				m.SetTitle(txt.Title(loc.Name()), SrcAuto)
			} else if len(loc.Name()) > 20 || len(loc.City()) > 16 || strings.Contains(loc.Name(), loc.City()) {
				m.SetTitle(fmt.Sprintf("%s / %s", loc.Name(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", loc.Name(), loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if loc.City() != "" && loc.CountryName() != "" {
			if len(loc.City()) > 20 {
				m.SetTitle(fmt.Sprintf("%s / %s", loc.City(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", loc.City(), loc.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			}
		}
	} else if m.PlaceLoaded() {
		knownLocation = true

		if names != "" {
			log.Debugf("photo: %s title based on %s (%s)", m.String(), english.Plural(len(people), "person", "people"), clean.Log(names))

			if l := len([]rune(names)); l > 35 {
				m.SetTitle(names, SrcAuto)
			} else if l > 20 && (m.Place.NoCity() || m.Place.LongCity()) {
				m.SetTitle(fmt.Sprintf("%s / %s", names, m.TakenAt.Format("2006")), SrcAuto)
			} else if l > 20 {
				m.SetTitle(fmt.Sprintf("%s / %s", names, m.Place.City()), SrcAuto)
			} else if m.Place.NoCity() || m.Place.LongCity() {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", names, m.Place.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", names, m.Place.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if title := labels.Title(fileTitle); title != "" {
			log.Debugf("photo: %s title based on label %s", m.String(), clean.Log(title))
			if m.Place.NoCity() || m.Place.LongCity() || m.Place.CityContains(title) {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), m.Place.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", txt.Title(title), m.Place.City(), m.TakenAt.Format("2006")), SrcAuto)
			}
		} else if m.Place.City() != "" && m.Place.CountryName() != "" {
			if len(m.Place.City()) > 20 {
				m.SetTitle(fmt.Sprintf("%s / %s", m.Place.City(), m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(fmt.Sprintf("%s / %s / %s", m.Place.City(), m.Place.CountryName(), m.TakenAt.Format("2006")), SrcAuto)
			}
		}
	}

	if !knownLocation || m.NoTitle() {
		if names != "" {
			if len([]rune(names)) <= 35 && m.TakenSrc != SrcAuto {
				m.SetTitle(fmt.Sprintf("%s / %s", names, m.TakenAt.Format("2006")), SrcAuto)
			} else {
				m.SetTitle(names, SrcAuto)
			}
		} else if fileTitle != "" && len(fileTitle) <= 20 && !m.TakenAtLocal.IsZero() && m.TakenSrc != SrcAuto {
			m.SetTitle(fmt.Sprintf("%s / %s", fileTitle, m.TakenAtLocal.Format("2006")), SrcAuto)
		} else if fileTitle != "" {
			m.SetTitle(fileTitle, SrcAuto)
		} else {
			m.SetTitle(UnknownTitle, SrcAuto)
		}
	}

	// Log changes.
	if m.PhotoTitle != oldTitle {
		log.Debugf("photo: %s has new title %s [%s]", m.String(), clean.Log(m.PhotoTitle), time.Since(start))
	}

	return nil
}

// GenerateAndSaveTitle updates the photo title and saves it.
func (m *Photo) GenerateAndSaveTitle() error {
	if !m.HasID() {
		return fmt.Errorf("photo id is missing")
	}

	m.PhotoFaces = m.FaceCount()

	labels := m.ClassifyLabels()

	m.UpdateDateFields()

	if err := m.GenerateTitle(labels); err != nil {
		log.Info(err)
	}

	details := m.GetDetails()

	w := txt.UniqueWords(txt.Words(details.Keywords))
	w = append(w, labels.Keywords()...)
	details.Keywords = strings.Join(txt.UniqueWords(w), ", ")

	if err := m.IndexKeywords(); err != nil {
		log.Errorf("photo: %s", err.Error())
	}

	if err := m.Save(); err != nil {
		return err
	}

	return nil
}

// FileTitle returns a photo title based on the file name and/or path.
func (m *Photo) FileTitle() string {
	// Generate title based on photo name, if not generated:
	if !fs.IsGenerated(m.PhotoName) {
		if title := txt.FileTitle(m.PhotoName); title != "" {
			return title
		}
	}

	// Generate title based on original file name, if any:
	if m.OriginalName != "" {
		if title := txt.FileTitle(m.OriginalName); !fs.IsGenerated(m.OriginalName) && title != "" {
			return title
		} else if title := txt.FileTitle(filepath.Dir(m.OriginalName)); title != "" {
			return title
		}
	}

	// Generate title based on photo path, if any:
	if m.PhotoPath != "" && !fs.IsGenerated(m.PhotoPath) {
		return txt.FileTitle(m.PhotoPath)
	}

	return ""
}

// UpdateTitleLabels updates the labels assigned based on the photo title.
func (m *Photo) UpdateTitleLabels() error {
	if m == nil {
		return nil
	} else if m.PhotoTitle == "" {
		return nil
	} else if SrcPriority[m.TitleSrc] < SrcPriority[SrcName] {
		return nil
	}

	keywords := txt.UniqueKeywords(m.PhotoTitle)

	var labelIds []uint

	for _, w := range keywords {
		if label, err := FindLabel(w, true); err == nil {
			if label.Skip() {
				continue
			}

			labelIds = append(labelIds, label.ID)
			FirstOrCreatePhotoLabel(NewPhotoLabel(m.ID, label.ID, 10, classify.SrcTitle))
		}
	}

	return Db().Where("label_src = ? AND photo_id = ? AND label_id NOT IN (?)", classify.SrcTitle, m.ID, labelIds).Delete(&PhotoLabel{}).Error
}
