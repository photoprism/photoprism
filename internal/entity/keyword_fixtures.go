package entity

type KeywordMap map[string]Keyword

func (m KeywordMap) Get(name string) Keyword {
	if result, ok := m[name]; ok {
		return result
	}

	return *NewKeyword(name)
}

func (m KeywordMap) Pointer(name string) *Keyword {
	if result, ok := m[name]; ok {
		return &result
	}

	return NewKeyword(name)
}

var KeywordFixtures = KeywordMap{
	"bridge": {
		ID:      1000000,
		Keyword: "bridge",
		Skip:    false,
	},
	"beach": {
		ID:      1000001,
		Keyword: "beach",
		Skip:    false,
	},
	"flower": {
		ID:      1000002,
		Keyword: "flower",
		Skip:    false,
	},
}

// CreateKeywordFixtures inserts known entities into the database for testing.
func CreateKeywordFixtures() {
	for _, entity := range KeywordFixtures {
		Db().Create(&entity)
	}
}
