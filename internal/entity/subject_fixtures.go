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
		SubjectUID:         "jqu0xs11qekk9jx8",
		SubjectSlug:        "john-doe",
		SubjectName:        "John Doe",
		SubjectType:        SubjectPerson,
		SubjectSrc:         SrcManual,
		Favorite:           true,
		Private:            false,
		Hidden:             false,
		SubjectDescription: "Subject Description",
		SubjectNotes:       "Short Note",
		MetadataJSON:       []byte(""),
		PhotoCount:         1,
		CreatedAt:          Timestamp(),
		UpdatedAt:          Timestamp(),
		DeletedAt:          nil,
	},
	"joe-biden": Subject{
		SubjectUID:         "jqy3y652h8njw0sx",
		SubjectSlug:        "joe-biden",
		SubjectName:        "Joe Biden",
		SubjectType:        SubjectPerson,
		SubjectSrc:         SrcMarker,
		Favorite:           false,
		Private:            false,
		Hidden:             false,
		SubjectDescription: "",
		SubjectNotes:       "",
		MetadataJSON:       []byte(""),
		PhotoCount:         1,
		CreatedAt:          Timestamp(),
		UpdatedAt:          Timestamp(),
		DeletedAt:          nil,
	},
}

// CreateSubjectFixtures inserts known entities into the database for testing.
func CreateSubjectFixtures() {
	for _, entity := range SubjectFixtures {
		Db().Create(&entity)
	}
}
