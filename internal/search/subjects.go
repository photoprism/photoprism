package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// Subjects searches subjects and returns them.
func Subjects(f form.SubjectSearch) (results SubjectResults, err error) {
	if err := f.ParseQueryString(); err != nil {
		return results, err
	}

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
		s = s.Where(fmt.Sprintf("%s.subj_uid IN (?)", entity.Subject{}.TableName()), strings.Split(f.ID, txt.Or))

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
		s = s.Where("subj_type IN (?)", strings.Split(f.Type, txt.Or))
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

// SubjectUIDs finds subject UIDs matching the search string, and removes names from the remaining query.
func SubjectUIDs(s string) (result []string, names []string, remaining string) {
	if s == "" {
		return result, names, s
	}

	type Matches struct {
		SubjUID   string
		SubjName  string
		SubjAlias string
	}

	var matches []Matches

	wheres := LikeAllNames(Cols{"subj_name", "subj_alias"}, s)

	if len(wheres) == 0 {
		return result, names, s
	}

	remaining = s

	for _, where := range wheres {
		var subj []string

		stmt := Db().Model(entity.Subject{})
		stmt = stmt.Where("?", gorm.Expr(where))

		if err := stmt.Scan(&matches).Error; err != nil {
			log.Errorf("search: %s while finding subjects", err)
		} else if len(matches) == 0 {
			continue
		}

		for _, m := range matches {
			subj = append(subj, m.SubjUID)
			names = append(names, m.SubjName)

			for _, r := range txt.Words(strings.ToLower(m.SubjName)) {
				if len(r) > 1 {
					remaining = strings.ReplaceAll(remaining, r, "")
				}
			}

			for _, r := range txt.Words(strings.ToLower(m.SubjAlias)) {
				if len(r) > 1 {
					remaining = strings.ReplaceAll(remaining, r, "")
				}
			}
		}

		result = append(result, strings.Join(subj, txt.Or))
	}

	return result, names, txt.NormalizeQuery(remaining)
}
