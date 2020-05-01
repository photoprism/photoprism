package entity

var CategoryFixtures = map[string]Category{
	"1": {
		LabelID:    1000001,
		CategoryID: 1000000,
		Label:      &LabelFixtureFlower,
		Category:   &LabelFixtureLandscape,
	},
}

// CreateCategoryFixtures inserts known entities into the database for testing.
func CreateCategoryFixtures() {
	for _, entity := range KeywordFixtures {
		Db().Create(&entity)
	}
}
