package entity

import (
	"fmt"
)

// Subjects represents a list of subjects.
type Subjects []Subject

// Delete (soft) deletes all subjects.
func (m Subjects) Delete() error {
	for _, subj := range m {
		if err := subj.Delete(); err != nil {
			return err
		}
	}

	return nil
}

// OrphanPeople returns unused subjects.
func OrphanPeople() (Subjects, error) {
	orphans := Subjects{}

	err := Db().
		Where("subj_type = ?", SubjPerson).
		Where(fmt.Sprintf("subj_uid NOT IN (SELECT DISTINCT subj_uid FROM %s)", Marker{}.TableName())).
		Find(&orphans).Error

	return orphans, err
}

// DeleteOrphanPeople finds and (soft) deletes all unused people.
func DeleteOrphanPeople() (count int, err error) {
	subj, err := OrphanPeople()

	if err != nil {
		return 0, err
	}

	return len(subj), subj.Delete()
}
