package query

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/convert"
)

// People returns the sorted names of the first 2000 people.
func People() (people entity.People, err error) {
	people = make(entity.People, 0)
	err = UnscopedDb().
		Table(entity.Subject{}.TableName()).
		Select("subj_uid, subj_name, subj_alias, subj_favorite, subj_hidden").
		Where("deleted_at IS NULL AND subj_type = ?", entity.SubjPerson).
		Order("subj_name").
		Limit(2000).Offset(0).
		Scan(&people).Error

	return people, err
}

// PeopleCount returns the total number of people in the index.
func PeopleCount() (count int, err error) {
	countData := int64(0)
	err = Db().
		Table(entity.Subject{}.TableName()).
		Where("deleted_at IS NULL").
		Where("subj_hidden = FALSE").
		Where("subj_type = ?", entity.SubjPerson).
		Count(&countData).Error

	return convert.SafeInt64toint(countData), err
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

// RemoveOrphanSubjects permanently removes dangling marker subjects from the index.
//
// A live verified person is kept: re-clustering leaves them unreferenced by design, and the row is
// what makes the same name comparable across runs. A soft-deleted one is collected whatever the flag
// says, since this is also the garbage collection for the tombstone MergeWith leaves.
func RemoveOrphanSubjects() (removed int64, err error) {

	// Gather all the uid's to be removed
	results := []string{}
	UnscopedDb().
		Model(&entity.Subject{}).
		Select("subj_uid").
		Where("subj_src = ?", entity.SrcMarker).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Face{}.TableName())).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Marker{}.TableName())).
		Scan(&results)

	res := UnscopedDb().
		Where("subj_src = ?", entity.SrcMarker).
		Where("(deleted_at IS NOT NULL OR verified = ?)", false).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Face{}.TableName())).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Subject{})

	if res.Error == nil {
		// Remove from the cache.  This is because BulkDelete appears to trigger a single AfterDelete that doesn't have the SubjUID in it.
		entity.SubjNames.BulkRemove(results)
	}

	return res.RowsAffected, res.Error
}

// CreateMarkerSubjects adds and references known marker subjects.
func CreateMarkerSubjects() (affected int64, err error) {
	var markers entity.Markers

	if err = Db().
		Where("subj_uid = '' AND marker_name <> '' AND subj_src <> ?", entity.SrcAuto).
		Where("marker_invalid = FALSE AND marker_type = ?", entity.MarkerFace).
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
			log.Errorf("faces: invalid subject %s", clean.Log(m.MarkerName))
			continue
		} else if subj = entity.FirstOrCreateSubject(subj); subj == nil {
			log.Errorf("faces: failed to add subject %s", clean.Log(m.MarkerName))
			continue
		} else {
			affected++
		}

		name = m.MarkerName
		m.SubjUID = subj.SubjUID
		m.MarkerReview = false

		if err = m.Updates(entity.Values{"subj_uid": m.SubjUID, "marker_review": m.MarkerReview}); err != nil {
			return affected, err
		}

		if m.FaceID == "" {
			continue
		} else if err = Db().Model(&entity.Face{}).Where("id = ? AND subj_uid = ''", m.FaceID).Update("subj_uid", m.SubjUID).Error; err != nil {
			return affected, err
		}
	}

	return affected, err
}
