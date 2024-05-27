package entity

type SubjectMap map[string]Subject

func (m SubjectMap) Get(name string) Subject {
	if result, ok := m[name]; ok {
		return result
	}

	return Subject{}
}

func (m SubjectMap) Pointer(name string) *Subject {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Subject{}
}

var SubjectFixtures = SubjectMap{
	"john-doe": Subject{
		SubjUID:      "js6sg6b1qekk9jx8",
		SubjSlug:     "john-doe",
		SubjName:     "John Doe",
		SubjType:     SubjPerson,
		SubjSrc:      SrcManual,
		SubjFavorite: true,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "Subject Description",
		SubjNotes:    "Short Note",
		FileCount:    1,
		PhotoCount:   1,
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		DeletedAt:    nil,
	},
	"joe-biden": Subject{
		SubjUID:      "js6sg6b2h8njw0sx",
		SubjSlug:     "joe-biden",
		SubjName:     "Joe Biden",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "",
		SubjNotes:    "",
		FileCount:    1,
		PhotoCount:   1,
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		DeletedAt:    nil,
	},
	"dangling": Subject{
		SubjUID:      "js6sg6b1h1njaaaa",
		SubjSlug:     "dangling-subject",
		SubjName:     "Dangling Subject",
		SubjAlias:    "Powell",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "",
		SubjNotes:    "",
		Thumb:        "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
		FileCount:    0,
		PhotoCount:   0,
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		DeletedAt:    nil,
	},
	"jane-doe": Subject{
		SubjUID:      "js6sg6b1h1njaaab",
		SubjSlug:     "jane-doe",
		SubjName:     "Jane Doe",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "",
		SubjNotes:    "",
		FileCount:    3,
		PhotoCount:   2,
		CreatedAt:    Now().AddDate(0, 0, 1),
		UpdatedAt:    Now(),
		DeletedAt:    nil,
	},
	"actress-1": Subject{
		SubjUID:      "js6sg6b1h1njaaac",
		SubjSlug:     "actress-a",
		SubjName:     "Actress A",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjNotes:    "",
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		DeletedAt:    nil,
	},
	"actor-1": Subject{
		SubjUID:      "js6sg6b1h1njaaad",
		SubjSlug:     "actor-a",
		SubjName:     "Actor A",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjNotes:    "",
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		DeletedAt:    nil,
	},
}

// CreateSubjectFixtures inserts known entities into the database for testing.
func CreateSubjectFixtures() {
	for _, entity := range SubjectFixtures {
		Db().Create(&entity)
	}
}
