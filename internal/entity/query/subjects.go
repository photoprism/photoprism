package query

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// People returns the sorted names of the first 2000 people.
func People() (people entity.People, err error) {
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
	err = Db().
		Table(entity.Subject{}.TableName()).
		Where("deleted_at IS NULL").
		Where("subj_hidden = FALSE").
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

// RemoveOrphanSubjects permanently removes dangling marker subjects from the index. A live person
// marked as Verified is kept, so the name stays comparable across re-clustering runs; a soft-deleted
// one is collected whatever the flag says, as this also removes the tombstone MergeWith leaves.
func RemoveOrphanSubjects() (removed int64, err error) {
	res := UnscopedDb().
		Where("subj_src = ?", entity.SrcMarker).
		Where("(deleted_at IS NOT NULL OR verified = ?)", false).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Face{}.TableName())).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT subj_uid FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Subject{})

	return res.RowsAffected, res.Error
}

// CreateMarkerSubjects adds and references known marker subjects, and returns how many names it linked
// markers to and how many XMP markers it linked to existing people, also on an error. A name from a
// source that may not name its person, such as XMP, is linked only to an existing person.
func CreateMarkerSubjects() (subjects, linked int64, err error) {
	var markers entity.Markers

	if err = Db().
		Where("subj_uid = '' AND marker_name <> '' AND subj_src <> ?", entity.SrcAuto).
		Where("marker_invalid = FALSE AND marker_type = ?", entity.MarkerFace).
		// Sorted by source within a name, so a person another source creates exists before an XMP
		// marker of the same name looks for it.
		Order("LOWER(marker_name), subj_src").
		Find(&markers).Error; err != nil {
		return subjects, linked, err
	} else if len(markers) == 0 {
		return subjects, linked, nil
	}

	// People resolved in this pass, keyed by the lowercase name, so case variants of one name are
	// resolved once and counted once, when the first marker is linked.
	resolved := make(map[string]*entity.Subject)
	counted := make(map[string]bool)

	for _, m := range markers {
		// A name from a source that may not name its person, such as XMP, is linked only to a person
		// who already exists, and never names the cluster.
		if !m.SourceNamesFace() {
			if found := entity.FindSubjectByName(m.MarkerName, false); found == nil || found.Deleted() || !found.IsPerson() {
				continue
			} else if err = m.Updates(entity.Values{"subj_uid": found.SubjUID, "marker_name": found.SubjName, "marker_review": false}); err != nil {
				return subjects, linked, err
			}

			linked++
			continue
		}

		key := strings.ToLower(clean.Name(m.MarkerName))
		subj := resolved[key]

		if subj != nil {
			// Resolved already.
		} else if subj = entity.NewSubject(m.MarkerName, entity.SubjPerson, entity.SrcMarker); subj == nil {
			log.Errorf("faces: invalid subject %s", clean.Log(m.MarkerName))
			continue
		} else if subj = entity.FirstOrCreateSubject(subj); subj == nil {
			log.Errorf("faces: failed to add subject %s", clean.Log(m.MarkerName))
			continue
		} else {
			resolved[key] = subj
		}

		m.SubjUID = subj.SubjUID
		m.MarkerReview = false

		if err = m.Updates(entity.Values{"subj_uid": m.SubjUID, "marker_review": m.MarkerReview}); err != nil {
			return subjects, linked, err
		} else if !counted[key] {
			counted[key] = true
			subjects++
		}

		if m.FaceID == "" {
			continue
		} else if err = Db().Model(&entity.Face{}).Where("id = ? AND subj_uid = ''", m.FaceID).Update("subj_uid", m.SubjUID).Error; err != nil {
			return subjects, linked, err
		}
	}

	return subjects, linked, err
}
