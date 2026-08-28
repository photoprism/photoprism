package entity

// UnknownFace can be used as a placeholder for unknown faces.
var UnknownFace = Face{
	ID:            UnknownID,
	FaceSrc:       SrcDefault,
	MatchedAt:     TimeStamp(),
	SubjUID:       "",
	EmbeddingJSON: []byte{},
}

type FaceMap map[string]Face

func (m FaceMap) Get(name string) Face {
	if result, ok := m[name]; ok {
		return result
	}

	return UnknownFace
}

func (m FaceMap) Pointer(name string) *Face {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UnknownFace
}

var FaceFixtures = FaceMap{
	"john-doe": Face{
		ID:           "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK",
		SubjUID:      SubjectFixtures.Get("john-doe").SubjUID,
		FaceSrc:      SrcAuto,
		SampleRadius: 0.8,
		Samples:      5,
		Collisions:   1,
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
	},
	"unknown": Face{
		ID:           "IW2P73ISBCUFPIAWSIOZKRDCHHFHC35S",
		SubjUID:      "",
		FaceSrc:      SrcAuto,
		SampleRadius: 0,
		Samples:      1,
		Collisions:   0,
		MatchedAt:    &editTime,
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
	},
	"joe-biden": Face{
		ID:      "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG6",
		SubjUID: SubjectFixtures.Get("joe-biden").SubjUID,
		FaceSrc: SrcManual,
		// Deliberately above face.ClusterRadius, which no calibration in effect can
		// produce: this row stands in for a cluster written under an earlier one, and
		// proves the radius is clamped where it is read rather than only where it is set.
		SampleRadius:    2,
		Samples:         33,
		Collisions:      0,
		CollisionRadius: 0,
		CreatedAt:       Now(),
		UpdatedAt:       Now(),
	},
	"jane-doe": Face{
		ID:              "VF7ANLDET2BKZNT4VQWJMMC6HBEFDOG7",
		SubjUID:         SubjectFixtures.Get("jane-doe").SubjUID,
		FaceSrc:         SrcManual,
		SampleRadius:    0.2849559839760571,
		Samples:         3,
		Collisions:      0,
		CollisionRadius: 0,
		CreatedAt:       Now(),
		UpdatedAt:       Now(),
	},
	"fa-gr": Face{
		ID:              "TOSCDXCS4VI3PGIUTCNIQCNI6HSFXQVZ",
		SubjUID:         "",
		FaceSrc:         "",
		SampleRadius:    0.3335191224530258,
		Samples:         4,
		Collisions:      0,
		CollisionRadius: 0,
		CreatedAt:       Now(),
		UpdatedAt:       Now(),
	},
	"actress-1": Face{
		ID:              "GMH5NISEEULNJL6RATITOA3TMZXMTMCI",
		SubjUID:         SubjectFixtures.Get("actress-1").SubjUID,
		FaceSrc:         "",
		SampleRadius:    0.27852392873736237,
		Samples:         4,
		Collisions:      0,
		CollisionRadius: 0,
		CreatedAt:       Now(),
		UpdatedAt:       Now(),
	},
	"actor-1": Face{
		ID:              "PI6A2XGOTUXEFI7CBF4KCI5I2I3JEJHS",
		SubjUID:         SubjectFixtures.Get("actor-1").SubjUID,
		FaceSrc:         "",
		SampleRadius:    0.3239983399779298,
		Samples:         4,
		Collisions:      0,
		CollisionRadius: 0,
		CreatedAt:       Now(),
		UpdatedAt:       Now(),
	},
}

// CreateFaceFixtures inserts known entities into the database for testing.
func CreateFaceFixtures() {
	GenerateFaceFixtureVectors()

	for _, entity := range FaceFixtures {
		fixtureDb().Create(&entity)
	}
}

// ReopenForTest reopens a cluster the way a collision does, so tests outside this package can
// build the state a matching pass has to recognize on its way out.
func (m *Face) ReopenForTest() {
	m.reopen()
}
