package entity

import "sync"

// SubjNames is a uid/name (reverse) lookup map
var SubjNames = NewStringMap(nil)

func init() {
	onReady = append(onReady, initSubjNames)
}

// initSubjNames initializes the subject uid/name (reverse) lookup table.
func initSubjNames() {
	var results KeyValues

	// Fetch subjects from the database.
	if err := UnscopedDb().Model(Subject{}).Select("subj_uid AS k, subj_name AS v").
		Scan(&results).Error; err != nil {
		log.Warnf("subjects: %s (init lookup)", err)
	} else {
		SubjNames = NewStringMap(results.Strings())
	}
}

// subjNamesMutex serializes changes made by setSubjName, so a retraction and the following set
// are not interleaved with another change of the same entry.
var subjNamesMutex = sync.Mutex{}

// setSubjName maps a subject uid to its name in SubjNames, retracting the uid from the reverse
// lookup of a different previous name so that name no longer resolves to it.
func setSubjName(uid, name string) {
	names := SubjNames

	if names.Unchanged(uid, name) {
		return
	}

	subjNamesMutex.Lock()
	defer subjNamesMutex.Unlock()

	if prev := names.Get(uid); prev != "" && prev != name {
		names.Unset(uid)
	}

	names.Set(uid, name)
}
