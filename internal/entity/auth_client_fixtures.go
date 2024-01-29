package entity

import (
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/unix"
)

type ClientMap map[string]Client

func (m ClientMap) Get(name string) Client {
	if result, ok := m[name]; ok {
		return result
	}

	return Client{}
}

func (m ClientMap) Pointer(name string) *Client {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Client{}
}

var ClientFixtures = ClientMap{
	"alice": {
		ClientUID:    "cs5gfen1bgxz7s9i",
		UserUID:      UserFixtures.Pointer("alice").UserUID,
		UserName:     UserFixtures.Pointer("alice").UserName,
		user:         UserFixtures.Pointer("alice"),
		ClientName:   "Alice",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientConfidential,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "*",
		AuthExpires:  unix.Day,
		AuthTokens:   -1,
		AuthEnabled:  true,
		LastActive:   0,
	},
	"bob": {
		ClientUID:    "cs5gfsvbd7ejzn8m",
		UserUID:      UserFixtures.Pointer("bob").UserUID,
		UserName:     UserFixtures.Pointer("bob").UserName,
		user:         UserFixtures.Pointer("bob"),
		ClientName:   "Bob",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientPublic,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "*",
		AuthExpires:  0,
		AuthTokens:   -1,
		AuthEnabled:  false,
		LastActive:   0,
	},
	"metrics": {
		ClientUID:    "cs5cpu17n6gj2qo5",
		UserUID:      "",
		UserName:     "",
		user:         nil,
		ClientName:   "Monitoring",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientConfidential,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "metrics",
		AuthExpires:  unix.Hour,
		AuthTokens:   2,
		AuthEnabled:  true,
		LastActive:   0,
	},
	"Unknown": {
		ClientUID:    "cs5cpu17n6gj2jh6",
		UserUID:      "",
		UserName:     "",
		user:         nil,
		ClientName:   "Unknown",
		ClientRole:   acl.RoleNone.String(),
		ClientType:   authn.ClientUnknown,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodUnknown.String(),
		AuthScope:    "*",
		AuthExpires:  unix.Hour,
		AuthTokens:   2,
		AuthEnabled:  true,
		LastActive:   0,
	},
	"deleted": {
		ClientUID:    "cs5cpu17n6gj2gf7",
		UserUID:      "",
		UserName:     "",
		user:         nil,
		ClientName:   "Deleted Monitoring",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientConfidential,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "metrics",
		AuthExpires:  unix.Hour,
		AuthTokens:   2,
		AuthEnabled:  true,
		LastActive:   0,
		DeletedAt:    TimePointer(),
	},
	"analytics": {
		ClientUID:    "cs7pvt5h8rw9aaqj",
		UserUID:      "",
		UserName:     "",
		user:         nil,
		ClientName:   "Analytics",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientConfidential,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "statistics",
		AuthExpires:  unix.Hour,
		AuthTokens:   2,
		AuthEnabled:  true,
		LastActive:   0,
	},
}

// CreateClientFixtures inserts known entities into the database for testing.
func CreateClientFixtures() {
	for _, entity := range ClientFixtures {
		Db().Create(&entity)
	}
}
