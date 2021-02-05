package query

// PurgeUnusedCountries removes countries without any photos.
func PurgeUnusedCountries() error {
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM countries WHERE id NOT IN (SELECT photo_country FROM photos)`).Error
	}
}
