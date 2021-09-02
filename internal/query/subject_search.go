package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/capture"
)

// SubjectResult represents a subject search result.
type SubjectResult struct {
	SubjectUID   string `json:"UID"`
	SubjectType  string `json:"Type"`
	SubjectSlug  string `json:"Slug"`
	SubjectName  string `json:"Name"`
	SubjectAlias string `json:"Alias,omitempty"`
	Thumb        string `json:"Thumb,omitempty"`
	Favorite     bool   `json:"Favorite,omitempty"`
	Private      bool   `json:"Private,omitempty"`
	Excluded     bool   `json:"Excluded,omitempty"`
	FileCount    int    `json:"Files,omitempty"`
}

// SubjectResults represents subject search results.
type SubjectResults []SubjectResult

// SubjectSearch searches subjects and returns them.
func SubjectSearch(f form.SubjectSearch) (results SubjectResults, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

	defer log.Debug(capture.Time(time.Now(), fmt.Sprintf("subjects: search %s", form.Serialize(f, true))))

	// Base query.
	s := UnscopedDb().Table(entity.Subject{}.TableName()).
		Select("subject_uid, subject_slug, subject_name, subject_alias, subject_type, thumb, favorite, private, excluded, file_count")

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Set sort order.
	switch f.Order {
	case "count":
		s = s.Order("file_count DESC")
	default:
		s = s.Order("subject_name")
	}

	if f.ID != "" {
		s = s.Where("subject_uid IN (?)", strings.Split(f.ID, Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		for _, where := range LikeAnyWord("subject_name", f.Query) {
			s = s.Where("(?)", gorm.Expr(where))
		}
	}

	if f.Type != "" {
		s = s.Where("subject_type IN (?)", strings.Split(f.Type, Or))
	}

	if f.Favorite {
		s = s.Where("favorite = 1")
	}

	if f.Private {
		s = s.Where("private = 1")
	}

	if f.Excluded {
		s = s.Where("excluded = 1")
	}

	s = s.Where("deleted_at IS NULL")

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
