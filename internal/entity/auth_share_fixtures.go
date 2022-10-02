package entity

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

type ShareMap map[string]Share

// Get returns a fixture for use in tests.
func (m ShareMap) Get(name string) Share {
	if result, ok := m[name]; ok {
		return result
	}

	return Share{}
}

// Pointer returns a fixture pointer for use in tests.
func (m ShareMap) Pointer(name string) *Share {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Share{}
}

// ShareFixtures specifies fixtures for use in tests.
var ShareFixtures = ShareMap{
	"AliceAlbum": {
		UserUID:   "uqxetse3cy5eo9z2",
		ShareUID:  "at9lxuqxpogaaba9",
		ExpiresAt: nil,
		Comment:   "The quick brown fox jumps over the lazy dog.",
		Perm:      PermShare,
		RefID:     rnd.RefID(SharePrefix),
		CreatedAt: TimeStamp(),
		UpdatedAt: TimeStamp(),
	},
}

// CreateShareFixtures creates the fixtures specified above.
func CreateShareFixtures() {
	for _, entity := range ShareFixtures {
		Db().Create(&entity)
	}
}
