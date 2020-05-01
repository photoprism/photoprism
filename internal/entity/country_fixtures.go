package entity

var CountryFixtures = map[string]Country{
	"apple-iphone-se": {
		ID:                 "de",
		CountrySlug:        "germany",
		CountryName:        "Germany",
		CountryDescription: "Country description",
		CountryNotes:       "Country Notes",
		CountryPhoto:       nil,
		CountryPhotoID:     0,
		New:                false,
	},
}

// CreateCountryFixtures inserts known entities into the database for testing.
func CreateCountryFixtures() {
	for _, entity := range CountryFixtures {
		Db().Create(&entity)
	}
}
