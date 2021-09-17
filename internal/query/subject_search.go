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
	SubjType     string `json:"Type"`
	SubjSlug     string `json:"Slug"`
	SubjName     string `json:"Name"`
	SubjAlias    string `json:"Alias"`
	SubjFavorite bool   `json:"Favorite"`
	SubjPrivate  bool   `json:"Private"`
	SubjExcluded bool   `json:"Excluded"`
	FileCount    int    `json:"Files"`
	Thumb        string `json:"Thumb"`
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
		Select("subj_uid, subj_slug, subj_name, subj_alias, subj_type, thumb, subj_favorite, subj_private, subj_excluded, file_count")

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
		s = s.Order("created_at DESC")
	case "relevance":
		s = s.Order("subj_favorite DESC, subj_name")
	default:
		s = s.Order("subj_favorite DESC, subj_name")
	}

	if f.ID != "" {
		s = s.Where("subj_uid IN (?)", strings.Split(f.ID, Or))

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
	s = s.Where("deleted_at IS NULL")

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
