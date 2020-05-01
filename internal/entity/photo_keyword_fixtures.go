package entity

var PhotoKeywordFixtures = map[string]PhotoKeyword{
	"1": {
		PhotoID:   1000004,
		KeywordID: 1000000,
	},
}

// CreatePhotoKeywordFixtures inserts known entities into the database for testing.
func CreatePhotoKeywordFixtures() {
	for _, entity := range PhotoKeywordFixtures {
		Db().Create(&entity)
	}
}
