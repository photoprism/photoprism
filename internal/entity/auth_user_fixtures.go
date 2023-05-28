package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

type UserMap map[string]User

// Get returns a user fixture for use in tests.
func (m UserMap) Get(name string) User {
	if result, ok := m[name]; ok {
		return result
	}

	return User{}
}

// Pointer returns a user fixture pointer for use in tests.
func (m UserMap) Pointer(name string) *User {
	if result, ok := m[name]; ok {
		return &result
	}

	return &User{}
}

// UserFixtures specifies user fixtures for use in tests.
var UserFixtures = UserMap{
	"alice": {
		ID:           5,
		UserUID:      "uqxetse3cy5eo9z2",
		UserName:     "alice",
		DisplayName:  "Alice",
		UserEmail:    "alice@example.com",
		UserRole:     acl.RoleAdmin.String(),
		AuthProvider: authn.ProviderLocal.String(),
		SuperAdmin:   true,
		CanLogin:     true,
		WebDAV:       true,
		CanInvite:    true,
		InviteToken:  GenerateToken(),
		UserSettings: &UserSettings{
			UITheme:     "",
			MapsStyle:   "",
			MapsAnimate: -1,
			UILanguage:  "",
			UITimeZone:  "",
		},
		UserDetails: &UserDetails{
			NickName:   "Lys",
			UserGender: GenderFemale,
		},
	},
	"bob": {
		ID:           7,
		UserUID:      "uqxc08w3d0ej2283",
		UserName:     "bob",
		DisplayName:  "Robert Rich",
		UserEmail:    "bob@example.com",
		UserRole:     acl.RoleAdmin.String(),
		AuthProvider: authn.ProviderLocal.String(),
		SuperAdmin:   false,
		CanLogin:     true,
		WebDAV:       true,
		CanInvite:    false,
		UserSettings: &UserSettings{
			UITheme:     "grayscale",
			MapsStyle:   "topographique",
			MapsAnimate: 6250,
			UILanguage:  "pt_BR",
			UITimeZone:  "",
		},
		UserDetails: &UserDetails{
			NickName:   "Bob",
			UserGender: GenderMale,
			BirthDay:   22,
			BirthMonth: 1,
			BirthYear:  1981,
		},
	},
	"friend": {
		ID:           8,
		UserUID:      "uqxqg7i1kperxvu7",
		UserName:     "friend",
		UserEmail:    "friend@example.com",
		UserRole:     acl.RoleAdmin.String(),
		AuthProvider: authn.ProviderLocal.String(),
		SuperAdmin:   false,
		DisplayName:  "Guy Friend",
		CanLogin:     true,
		WebDAV:       false,
		CanInvite:    false,
		UserSettings: &UserSettings{
			UITheme:     "gemstone",
			MapsStyle:   "hybrid",
			MapsAnimate: 0,
			UILanguage:  "es_US",
			UITimeZone:  "America/Los_Angeles",
		},
		UserDetails: &UserDetails{
			UserGender: GenderOther,
		},
	},
	"deleted": {
		ID:           10000008,
		UserUID:      "uqxqg7i1kperxvu8",
		UserName:     "deleted",
		UserEmail:    "",
		DisplayName:  "Deleted User",
		SuperAdmin:   false,
		UserRole:     acl.RoleVisitor.String(),
		AuthProvider: authn.ProviderLocal.String(),
		CanLogin:     false,
		WebDAV:       true,
		CanInvite:    false,
		DeletedAt:    TimePointer(),
		UserSettings: &UserSettings{
			UITheme:     "",
			MapsStyle:   "",
			MapsAnimate: 0,
			UILanguage:  "de",
			UITimeZone:  "",
		},
	},
	"unauthorized": {
		ID:           10000009,
		UserUID:      "uriku0138hqql4bz",
		UserName:     "jens.mander",
		UserEmail:    "jens.mander@microsoft.com",
		UserRole:     acl.RoleUnknown.String(),
		AuthProvider: authn.ProviderNone.String(),
		SuperAdmin:   false,
		DisplayName:  "Jens Mander",
		CanLogin:     true,
		WebDAV:       true,
		CanInvite:    false,
		UserSettings: &UserSettings{
			UITheme:     "",
			MapsStyle:   "",
			MapsAnimate: 0,
			UILanguage:  "de",
			UITimeZone:  "",
		},
		UserDetails: &UserDetails{
			UserGender: GenderMale,
		},
	},
	"fowler": {
		ID:           10000023,
		UserUID:      "urinotv3d6jedvlm",
		UserName:     "fowler",
		DisplayName:  "Martin Fowler",
		UserEmail:    "martin@fowler.org",
		UserRole:     acl.RoleAdmin.String(),
		AuthProvider: authn.ProviderLocal.String(),
		SuperAdmin:   false,
		CanLogin:     true,
		WebDAV:       true,
		CanInvite:    true,
		InviteToken:  GenerateToken(),
		UserSettings: &UserSettings{
			UITheme:     "custom",
			MapsStyle:   "invalid",
			MapsAnimate: -1,
			UILanguage:  "en",
			UITimeZone:  "UTC",
		},
	},
}

// CreateUserFixtures creates the user fixtures specified above
func CreateUserFixtures() {
	for _, entity := range UserFixtures {
		Db().Create(&entity)
	}
}
