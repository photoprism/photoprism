package entity

import (
	"github.com/photoprism/photoprism/internal/auth/acl"
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
		AuthMethod:   authn.MethodUndefined.String(),
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
		AuthMethod:   authn.MethodDefault.String(),
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
		AuthMethod:   authn.MethodDefault.String(),
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
		AuthMethod:   authn.MethodDefault.String(),
		CanLogin:     false,
		WebDAV:       true,
		CanInvite:    false,
		DeletedAt:    TimeStamp(),
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
		UserRole:     acl.RoleNone.String(),
		AuthProvider: authn.ProviderNone.String(),
		AuthMethod:   authn.MethodUndefined.String(),
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
		AuthMethod:   authn.MethodDefault.String(),
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
	"jane": {
		ID:           10000024,
		UserUID:      "usamyuogp49vd4lh",
		UserName:     "jane",
		DisplayName:  "Jane Dow",
		UserEmail:    "qa@example.com",
		UserRole:     acl.RoleAdmin.String(),
		AuthProvider: authn.ProviderLocal.String(),
		AuthMethod:   authn.Method2FA.String(),
		SuperAdmin:   false,
		CanLogin:     true,
		WebDAV:       true,
		CanInvite:    true,
		InviteToken:  GenerateToken(),
		UserSettings: &UserSettings{
			UITheme:     "default",
			MapsStyle:   "hybrid",
			MapsAnimate: 6250,
			UILanguage:  "en",
			UITimeZone:  "UTC",
		},
		UserDetails: &UserDetails{
			NickName:   "Jane",
			UserGender: GenderFemale,
			BirthDay:   23,
			BirthMonth: 5,
			BirthYear:  2001,
		},
	},
	"guest": {
		ID:           10000025,
		UserUID:      "usg73p55zwgr1gbq",
		UserName:     "guest",
		DisplayName:  "Guest User",
		UserEmail:    "guest@example.com",
		UserRole:     acl.RoleGuest.String(),
		AuthProvider: authn.ProviderOIDC.String(),
		AuthMethod:   authn.MethodDefault.String(),
		SuperAdmin:   false,
		CanLogin:     true,
		WebDAV:       false,
		CanInvite:    false,
		InviteToken:  "",
		UserSettings: &UserSettings{
			UITheme:     "default",
			MapsStyle:   "default",
			MapsAnimate: 6250,
			UILanguage:  "en",
			UITimeZone:  "UTC",
		},
		UserDetails: &UserDetails{
			NickName:   "Gustav",
			UserGender: GenderMale,
			BirthDay:   23,
			BirthMonth: 1,
			BirthYear:  1999,
		},
	},
	"no_local_auth": {
		ID:           10000026,
		UserUID:      "usg73p55zwgr1ytr",
		UserName:     "no_local_auth",
		DisplayName:  "Not Local",
		UserEmail:    "notlocal@example.com",
		UserRole:     acl.RoleGuest.String(),
		AuthProvider: authn.ProviderApplication.String(),
		AuthMethod:   authn.MethodDefault.String(),
		SuperAdmin:   false,
		CanLogin:     true,
		WebDAV:       false,
		CanInvite:    false,
		InviteToken:  "",
		UserSettings: &UserSettings{
			UITheme:     "default",
			MapsStyle:   "default",
			MapsAnimate: 6250,
			UILanguage:  "en",
			UITimeZone:  "UTC",
		},
	},
	"2fa": {
		ID:          10000027,
		UserUID:     "usg73p55zwgr1ojy",
		UserName:    "2fa",
		DisplayName: "2FA Enabled",
		UserEmail:   "2FA@example.com",

		UserRole:     acl.RoleAdmin.String(),
		AuthProvider: authn.ProviderLocal.String(),
		AuthMethod:   authn.Method2FA.String(),
		SuperAdmin:   false,
		CanLogin:     true,
		WebDAV:       false,
		CanInvite:    false,
		InviteToken:  "",
		UserSettings: &UserSettings{
			UITheme:     "default",
			MapsStyle:   "default",
			MapsAnimate: 6250,
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
