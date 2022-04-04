package entity

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
