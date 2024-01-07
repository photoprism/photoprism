package entity

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

type UserShareMap map[string]UserShare

// Get returns a fixture for use in tests.
func (m UserShareMap) Get(name string) UserShare {
	if result, ok := m[name]; ok {
		return result
	}

	return UserShare{}
}

// Pointer returns a fixture pointer for use in tests.
func (m UserShareMap) Pointer(name string) *UserShare {
	if result, ok := m[name]; ok {
		return &result
	}

	return &UserShare{}
}

// UserShareFixtures specifies fixtures for use in tests.
var UserShareFixtures = UserShareMap{
	"AliceAlbum": {
		UserUID:   "uqxetse3cy5eo9z2",
		ShareUID:  "as6sg6bxpogaaba9",
		ExpiresAt: nil,
		Comment:   "The quick brown fox jumps over the lazy dog.",
		Perm:      PermShare,
		RefID:     rnd.RefID(SharePrefix),
		CreatedAt: TimeStamp(),
		UpdatedAt: TimeStamp(),
	},
}

// CreateUserShareFixtures creates the fixtures specified above.
func CreateUserShareFixtures() {
	for _, entity := range UserShareFixtures {
		Db().Create(&entity)
	}
}
