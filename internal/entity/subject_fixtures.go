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
		SubjUID:      "jqu0xs11qekk9jx8",
		SubjSlug:     "john-doe",
		SubjName:     "John Doe",
		SubjType:     SubjPerson,
		SubjSrc:      SrcManual,
		SubjFavorite: true,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "Subject Description",
		SubjNotes:    "Short Note",
		MetadataJSON: []byte(""),
		FileCount:    1,
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		DeletedAt:    nil,
	},
	"joe-biden": Subject{
		SubjUID:      "jqy3y652h8njw0sx",
		SubjSlug:     "joe-biden",
		SubjName:     "Joe Biden",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "",
		SubjNotes:    "",
		MetadataJSON: []byte(""),
		FileCount:    1,
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		DeletedAt:    nil,
	},
	"dangling": Subject{
		SubjUID:      "jqy1y111h1njaaaa",
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
		MetadataJSON: []byte(""),
		FileCount:    0,
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		DeletedAt:    nil,
	},
	"jane-doe": Subject{
		SubjUID:      "jqy1y111h1njaaab",
		SubjSlug:     "jane-doe",
		SubjName:     "Jane Doe",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjExcluded: false,
		SubjBio:      "",
		SubjNotes:    "",
		MetadataJSON: []byte(""),
		FileCount:    3,
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		DeletedAt:    nil,
	},
	"actress-1": Subject{
		SubjUID:      "jqy1y111h1njaaac",
		SubjSlug:     "actress-a",
		SubjName:     "Actress A",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjNotes:    "",
		MetadataJSON: []byte(""),
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		DeletedAt:    nil,
	},
	"actor-1": Subject{
		SubjUID:      "jqy1y111h1njaaad",
		SubjSlug:     "actor-a",
		SubjName:     "Actor A",
		SubjType:     SubjPerson,
		SubjSrc:      SrcMarker,
		SubjFavorite: false,
		SubjPrivate:  false,
		SubjNotes:    "",
		MetadataJSON: []byte(""),
		CreatedAt:    TimeStamp(),
		UpdatedAt:    TimeStamp(),
		DeletedAt:    nil,
	},
}

// CreateSubjectFixtures inserts known entities into the database for testing.
func CreateSubjectFixtures() {
	for _, entity := range SubjectFixtures {
		Db().Create(&entity)
	}
}
