package entity

var CategoryFixtures = map[string]Category{
	"flower_landscape": {
		LabelID:    1000001,
		Label:      LabelFixtures.Pointer("flower"),
		CategoryID: 1000000,
		Category:   LabelFixtures.Pointer("landscape"),
	},
}

// CreateCategoryFixtures inserts known entities into the database for testing.
func CreateCategoryFixtures() {
	for _, entity := range CategoryFixtures {
		Db().Create(&entity)
	}
}
