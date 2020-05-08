package entity

var DescriptionFixtures = map[string]Description{
	"lake": {
		PhotoID:          1000000,
		PhotoDescription: "photo description",
		PhotoKeywords:    "nature, frog",
		PhotoNotes:       "notes",
		PhotoSubject:     "Lake",
		PhotoArtist:      "Hans",
		PhotoCopyright:   "copy",
		PhotoLicense:     "MIT",
	},
}
var DescriptionFixtureLake = DescriptionFixtures["lake"]

// CreateDescriptionFixtures inserts known entities into the database for testing.
func CreateDescriptionFixtures() {
	for _, entity := range DescriptionFixtures {
		Db().Create(&entity)
	}
}
