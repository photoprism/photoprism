package entity

import "github.com/photoprism/photoprism/pkg/authn"

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
		ClientUID:   "cs5gfen1bgxz7s9i",
		UserUID:     UserFixtures.Pointer("alice").UserUID,
		UserName:    UserFixtures.Pointer("alice").UserName,
		user:        UserFixtures.Pointer("alice"),
		ClientName:  "Alice",
		ClientType:  authn.ClientConfidential,
		ClientURL:   "",
		CallbackURL: "",
		AuthMethod:  authn.MethodOAuth2.String(),
		AuthScope:   "*",
		AuthExpires: UnixDay,
		AuthTokens:  -1,
		AuthEnabled: true,
		LastActive:  0,
	},
	"bob": {
		ClientUID:   "cs5gfsvbd7ejzn8m",
		UserUID:     UserFixtures.Pointer("bob").UserUID,
		UserName:    UserFixtures.Pointer("bob").UserName,
		user:        UserFixtures.Pointer("bob"),
		ClientName:  "Bob",
		ClientType:  authn.ClientWebDAV,
		ClientURL:   "",
		CallbackURL: "",
		AuthMethod:  authn.MethodBasic.String(),
		AuthScope:   "webdav files photos",
		AuthExpires: 0,
		AuthTokens:  -1,
		AuthEnabled: false,
		LastActive:  0,
	},
	"metrics": {
		ClientUID:   "cs5cpu17n6gj2qo5",
		UserUID:     "",
		UserName:    "",
		user:        nil,
		ClientName:  "Monitoring",
		ClientType:  authn.ClientConfidential,
		ClientURL:   "",
		CallbackURL: "",
		AuthMethod:  authn.MethodOAuth2.String(),
		AuthScope:   "metrics",
		AuthExpires: UnixHour,
		AuthTokens:  2,
		AuthEnabled: true,
		LastActive:  0,
	},
}

// CreateClientFixtures inserts known entities into the database for testing.
func CreateClientFixtures() {
	for _, entity := range ClientFixtures {
		Db().Create(&entity)
	}
}
