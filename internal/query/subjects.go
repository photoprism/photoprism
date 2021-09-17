package query

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/photoprism/photoprism/internal/entity"
)

// People returns the sorted names of the first 2000 people.
func People() (people entity.People, err error) {
	err = UnscopedDb().
		Table(entity.Subject{}.TableName()).
		Select("subj_uid, subj_name, subj_alias, subj_favorite ").
		Where("deleted_at IS NULL AND subj_type = ?", entity.SubjPerson).
		Order("subj_favorite, subj_name").
		Limit(2000).Offset(0).
		Scan(&people).Error

	return people, err
}

// PeopleCount returns the total number of people in the index.
func PeopleCount() (count int, err error) {
	err = Db().
		Table(entity.Subject{}.TableName()).
		Where("deleted_at IS NULL").
		Where("subj_type = ?", entity.SubjPerson).
		Count(&count).Error

	return count, err
}

// Subjects returns subjects from the index.
func Subjects(limit, offset int) (result entity.Subjects, err error) {
	stmt := Db()

	stmt = stmt.Order("subj_name").Limit(limit).Offset(offset)
	err = stmt.Find(&result).Error

	return result, err
}

// SubjectMap returns a map of subjects indexed by UID.
func SubjectMap() (result map[string]entity.Subject, err error) {
	result = make(map[string]entity.Subject)

	var subj entity.Subjects

	stmt := Db()

	if err = stmt.Find(&subj).Error; err != nil {
		return result, err
	}

	for _, s := range subj {
		result[s.SubjUID] = s
	}

	return result, err
}

// RemoveDanglingMarkerSubjects permanently deletes dangling marker subjects from the index.
func RemoveDanglingMarkerSubjects() (removed int64, err error) {
	res := UnscopedDb().
		Where("subj_src = ?", entity.SrcMarker).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Face{}.TableName())).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Subject{})

	return res.RowsAffected, res.Error
}

// CreateMarkerSubjects adds and references known marker subjects.
func CreateMarkerSubjects() (affected int64, err error) {
	var markers entity.Markers

	if err := Db().
		Where("subj_uid = '' AND marker_name <> '' AND subj_src <> ?", entity.SrcAuto).
		Where("marker_invalid = 0 AND marker_type = ?", entity.MarkerFace).
		Order("marker_name").
		Find(&markers).Error; err != nil {
		return affected, err
	} else if len(markers) == 0 {
		return affected, nil
	}

	var name string
	var subj *entity.Subject

	for _, m := range markers {
		if name == m.MarkerName && subj != nil {
			// Do nothing.
		} else if subj = entity.NewSubject(m.MarkerName, entity.SubjPerson, entity.SrcMarker); subj == nil {
			log.Errorf("faces: subject should not be nil - bug?")
			continue
		} else if subj = entity.FirstOrCreateSubject(subj); subj == nil {
			log.Errorf("faces: failed adding subject %s", txt.Quote(m.MarkerName))
			continue
		} else {
			affected++
		}

		name = m.MarkerName

		if err := m.Updates(entity.Values{"SubjUID": subj.SubjUID, "MarkerReview": false}); err != nil {
			return affected, err
		}

		if m.FaceID == "" {
			continue
		} else if err := Db().Model(&entity.Face{}).Where("id = ? AND subj_uid = ''", m.FaceID).Update("SubjUID", subj.SubjUID).Error; err != nil {
			return affected, err
		}
	}

	return affected, err
}

// SearchSubjUIDs finds subject UIDs matching the search string, and removes names from the remaining query.
func SearchSubjUIDs(s string) (result []string, names []string, remaining string) {
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

		result = append(result, strings.Join(subj, Or))
	}

	return result, names, NormalizeSearchQuery(remaining)
}
