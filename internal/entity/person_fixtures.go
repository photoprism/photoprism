package entity

type PersonMap map[string]Person

func (m PersonMap) Get(name string) Person {
	if result, ok := m[name]; ok {
		return result
	}

	return UnknownPerson
}

func (m PersonMap) Pointer(name string) *Person {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UnknownPerson
}

var PersonFixtures = PersonMap{
	"known_face": Person{
		ID:                2,
		PersonUID:         "rqu0xs11qekk9jx8",
		PersonSlug:        "john-doe",
		PersonName:        "John Doe",
		PersonSrc:         "xmp",
		PersonFavorite:    true,
		PersonPrivate:     false,
		PersonHidden:      false,
		PersonDescription: "Person Description",
		PersonNotes:       "Short Note",
		PersonMeta:        "",
		PhotoCount:        1,
		BirthYear:         2000,
		BirthMonth:        5,
		BirthDay:          22,
		PassedAway:        nil,
		CreatedAt:         Timestamp(),
		UpdatedAt:         Timestamp(),
		DeletedAt:         nil,
	},
}

// CreatePersonFixtures inserts known entities into the database for testing.
func CreatePersonFixtures() {
	for _, entity := range PersonFixtures {
		Db().Create(&entity)
	}
}
