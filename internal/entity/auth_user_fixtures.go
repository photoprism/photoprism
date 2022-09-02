package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

type UserMap map[string]User

func (m UserMap) Get(name string) User {
	if result, ok := m[name]; ok {
		return result
	}

	return User{}
}

func (m UserMap) Pointer(name string) *User {
	if result, ok := m[name]; ok {
		return &result
	}

	return &User{}
}

var UserFixtures = UserMap{
	"alice": {
		ID:          5,
		UserUID:     "uqxetse3cy5eo9z2",
		UserSlug:    "alice",
		Username:    "alice",
		Email:       "alice@example.com",
		UserRole:    acl.RoleAdmin.String(),
		SuperAdmin:  true,
		DisplayName: "Alice",
		CanLogin:    true,
		CanInvite:   true,
		InviteToken: rnd.GenerateToken(8),
	},
	"bob": {
		ID:          7,
		UserUID:     "uqxc08w3d0ej2283",
		UserSlug:    "bob",
		Username:    "bob",
		Email:       "bob@example.com",
		UserRole:    acl.RoleEditor.String(),
		SuperAdmin:  false,
		DisplayName: "Bob",
		CanLogin:    true,
		CanInvite:   false,
	},
	"friend": {
		ID:          8,
		UserUID:     "uqxqg7i1kperxvu7",
		UserSlug:    "friend",
		Username:    "friend",
		Email:       "friend@example.com",
		UserRole:    acl.RoleViewer.String(),
		SuperAdmin:  false,
		DisplayName: "Guy Friend",
		CanLogin:    true,
		CanInvite:   false,
	},
	"deleted": {
		ID:          10000008,
		UserUID:     "uqxqg7i1kperxvu8",
		UserSlug:    "deleted",
		Username:    "deleted",
		Email:       "",
		DisplayName: "Deleted User",
		SuperAdmin:  false,
		UserRole:    acl.RoleGuest.String(),
		CanLogin:    false,
		CanInvite:   false,
		DeletedAt:   &deleteTime,
	},
}

// CreateUserFixtures inserts known entities into the database for testing.
func CreateUserFixtures() {
	for _, entity := range UserFixtures {
		Db().Create(&entity)
	}
}
