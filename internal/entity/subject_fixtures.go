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
	"known_face": Subject{
		ID:                 2,
		SubjectUID:         "jqu0xs11qekk9jx8",
		SubjectSlug:        "john-doe",
		SubjectName:        "John Doe",
		SubjectSrc:         "xmp",
		Favorite:           true,
		Private:            false,
		Hidden:             false,
		SubjectDescription: "Subject Description",
		SubjectNotes:       "Short Note",
		SubjectJSON:        []byte(""),
		PhotoCount:         1,
		BirthYear:          2000,
		BirthMonth:         5,
		BirthDay:           22,
		PassedAway:         nil,
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
