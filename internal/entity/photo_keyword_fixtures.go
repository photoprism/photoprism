package entity

type PhotoKeywordMap map[string]PhotoKeyword

var PhotoKeywordFixtures = PhotoKeywordMap{
	"1": {
		PhotoID:   1000004,
		KeywordID: 1000000,
	},
	"2": {
		PhotoID:   1000001,
		KeywordID: 1000001,
	},
	"3": {
		PhotoID:   1000000,
		KeywordID: 1000000,
	},
	"4": {
		PhotoID:   1000003,
		KeywordID: 1000000,
	},
	"5": {
		PhotoID:   1000003,
		KeywordID: 1000003,
	},
	"6": {
		PhotoID:   1000023,
		KeywordID: 1000003,
	},
	"7": {
		PhotoID:   1000027,
		KeywordID: 1000004,
	},
	"8": {
		PhotoID:   1000030,
		KeywordID: 1000007,
	},
}

// CreatePhotoKeywordFixtures inserts known entities into the database for testing.
func CreatePhotoKeywordFixtures() {
	for _, entity := range PhotoKeywordFixtures {
		Db().Create(&entity)
	}
}
