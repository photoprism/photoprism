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
	SubjUID      string `json:"UID"`
	MarkerUID    string `json:"MarkerUID"`
	MarkerSrc    string `json:"MarkerSrc,omitempty"`
	SubjType     string `json:"Type"`
	SubjSlug     string `json:"Slug"`
	SubjName     string `json:"Name"`
	SubjAlias    string `json:"Alias"`
	SubjFavorite bool   `json:"Favorite"`
	SubjPrivate  bool   `json:"Private"`
	SubjExcluded bool   `json:"Excluded"`
	FileCount    int    `json:"FileCount"`
	FileHash     string `json:"FileHash"`
	CropArea     string `json:"CropArea"`
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
		Select(fmt.Sprintf("%s.*, m.file_hash, m.crop_area", entity.Subject{}.TableName()))

	// Join markers table for face thumbs.
	s = s.Joins(fmt.Sprintf("LEFT JOIN %s m ON m.marker_uid = %s.marker_uid", entity.Marker{}.TableName(), entity.Subject{}.TableName()))

	// Limit result count.
	if f.Count > 0 && f.Count <= MaxResults {
		s = s.Limit(f.Count).Offset(f.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(f.Offset)
	}

	// Set sort order.
	switch f.Order {
	case "name":
		s = s.Order("subj_name")
	case "count":
		s = s.Order("file_count DESC")
	case "added":
		s = s.Order(fmt.Sprintf("%s.created_at DESC", entity.Subject{}.TableName()))
	case "relevance":
		s = s.Order("subj_favorite DESC, subj_name")
	default:
		s = s.Order("subj_favorite DESC, subj_name")
	}

	if f.ID != "" {
		s = s.Where(fmt.Sprintf("%s.subj_uid IN (?)", entity.Subject{}.TableName()), strings.Split(f.ID, Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	if f.Query != "" {
		for _, where := range LikeAnyWord("subj_name", f.Query) {
			s = s.Where("(?)", gorm.Expr(where))
		}
	}

	if f.Files > 0 {
		s = s.Where("file_count >= ?", f.Files)
	}

	if f.Type != "" {
		s = s.Where("subj_type IN (?)", strings.Split(f.Type, Or))
	}

	if f.Favorite {
		s = s.Where("subj_favorite = 1")
	}

	if f.Private {
		s = s.Where("subj_private = 1")
	}

	if f.Excluded {
		s = s.Where("subj_excluded = 1")
	}

	// Omit deleted rows.
	s = s.Where(fmt.Sprintf("%s.deleted_at IS NULL", entity.Subject{}.TableName()))

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
