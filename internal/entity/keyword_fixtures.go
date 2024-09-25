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
	"kuh": {
		ID:      1000003,
		Keyword: "kuh",
		Skip:    false,
	},
	"actress": {
		ID:      1000004,
		Keyword: "actress",
		Skip:    false,
	},
	"%toss": {
		ID:      1000005,
		Keyword: "%toss",
		Skip:    false,
	},
	"ca%t": {
		ID:      1000006,
		Keyword: "ca%t",
		Skip:    false,
	},
	"magic%": {
		ID:      1000007,
		Keyword: "magic%",
		Skip:    false,
	},
	"&hogwarts": {
		ID:      1000008,
		Keyword: "&hogwarts",
		Skip:    false,
	},
	"love&trust": {
		ID:      1000009,
		Keyword: "love&trust",
		Skip:    false,
	},
	"countryside&": {
		ID:      10000010,
		Keyword: "countryside&",
		Skip:    false,
	},
	"'grandfather": {
		ID:      10000011,
		Keyword: "'grandfather",
		Skip:    false,
	},
	"grandma's": {
		ID:      10000012,
		Keyword: "grandma's",
		Skip:    false,
	},
	"cheescake'": {
		ID:      10000013,
		Keyword: "cheescake'",
		Skip:    false,
	},
	"*rating": {
		ID:      10000014,
		Keyword: "*rating",
		Skip:    false,
	},
	"three*four": {
		ID:      10000015,
		Keyword: "three*four",
		Skip:    false,
	},
	"tree*": {
		ID:      10000016,
		Keyword: "tree*",
		Skip:    false,
	},
	"|mystery": {
		ID:      10000017,
		Keyword: "|mystery",
		Skip:    false,
	},
	"run|stay": {
		ID:      10000018,
		Keyword: "run|stay",
		Skip:    false,
	},
	"pillow|": {
		ID:      10000019,
		Keyword: "pillow|",
		Skip:    false,
	},
	"1dish": {
		ID:      10000020,
		Keyword: "1dish",
		Skip:    false,
	},
	"nothing4you": {
		ID:      10000021,
		Keyword: "nothing4you",
		Skip:    false,
	},
	"joyx2": {
		ID:      10000022,
		Keyword: "joyx2",
		Skip:    false,
	},
	"\"electronics": {
		ID:      10000023,
		Keyword: "\"electronics",
		Skip:    false,
	},
	"sal\"mon": {
		ID:      10000024,
		Keyword: "sal\"mon",
		Skip:    false,
	},
	"fish\"": {
		ID:      10000025,
		Keyword: "fish\"",
		Skip:    false,
	},
}

// CreateKeywordFixtures inserts known entities into the database for testing.
func CreateKeywordFixtures() {
	for _, entity := range KeywordFixtures {
		Db().Create(&entity)
	}
}
