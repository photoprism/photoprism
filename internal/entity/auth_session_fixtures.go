package entity

import (
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/unix"
)

type SessionMap map[string]Session

func (m SessionMap) Get(name string) Session {
	if result, ok := m[name]; ok {
		return result
	}

	return Session{}
}

func (m SessionMap) Pointer(name string) *Session {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Session{}
}

var SessionFixtures = SessionMap{
	"alice": {
		authToken:   "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0",
		ID:          rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0"),
		RefID:       "sessxkkcabcd",
		SessTimeout: unix.Day * 3,
		SessExpires: unix.Now() + unix.Week,
		GrantType:   authn.GrantPassword.String(),
		user:        UserFixtures.Pointer("alice"),
		UserUID:     UserFixtures.Pointer("alice").UserUID,
		UserName:    UserFixtures.Pointer("alice").UserName,
	},
	"alice_token": {
		authToken:    "bb8658e779403ae524a188712470060f050054324a8b104e",
		ID:           rnd.SessionID("bb8658e779403ae524a188712470060f050054324a8b104e"),
		RefID:        "sess34q3hael",
		SessTimeout:  -1,
		SessExpires:  unix.Now() + unix.Day,
		AuthScope:    clean.Scope("*"),
		AuthProvider: authn.ProviderAccessToken.String(),
		AuthMethod:   authn.MethodDefault.String(),
		GrantType:    authn.GrantCLI.String(),
		ClientName:   "alice_token",
		LastActive:   -1,
		user:         UserFixtures.Pointer("alice"),
		UserUID:      UserFixtures.Pointer("alice").UserUID,
		UserName:     UserFixtures.Pointer("alice").UserName,
	},
	"alice_token_personal": {
		authToken:    "DIbS8T-uyGMe1-R3fmTv-vVaR35",
		ID:           rnd.SessionID("DIbS8T-uyGMe1-R3fmTv-vVaR35"),
		RefID:        "sess6ey1ykya",
		SessTimeout:  -1,
		SessExpires:  unix.Now() + unix.Day,
		AuthScope:    clean.Scope("*"),
		AuthProvider: authn.ProviderAccessToken.String(),
		AuthMethod:   authn.MethodDefault.String(),
		GrantType:    authn.GrantPassword.String(),
		ClientName:   "alice_token_personal",
		LastActive:   -1,
		user:         UserFixtures.Pointer("alice"),
		UserUID:      UserFixtures.Pointer("alice").UserUID,
		UserName:     UserFixtures.Pointer("alice").UserName,
	},
	"alice_token_webdav": {
		authToken:    "5d0rGx-EvsDnV-DcKtYY-HT1aWL",
		ID:           rnd.SessionID("5d0rGx-EvsDnV-DcKtYY-HT1aWL"),
		RefID:        "sesshjtgx8qt",
		SessTimeout:  -1,
		SessExpires:  unix.Now() + unix.Day,
		AuthScope:    clean.Scope("webdav"),
		AuthProvider: authn.ProviderAccessToken.String(),
		AuthMethod:   authn.MethodDefault.String(),
		GrantType:    authn.GrantPassword.String(),
		ClientName:   "alice_token_webdav",
		LastActive:   -1,
		user:         UserFixtures.Pointer("alice"),
		UserUID:      UserFixtures.Pointer("alice").UserUID,
		UserName:     UserFixtures.Pointer("alice").UserName,
	},
	"alice_token_scope": {
		authToken:     "778f0f7d80579a072836c65b786145d6e0127505194cc51e",
		ID:            rnd.SessionID("778f0f7d80579a072836c65b786145d6e0127505194cc51e"),
		RefID:         "sessjr0ge18d",
		SessTimeout:   0,
		SessExpires:   unix.Now() + unix.Day,
		AuthScope:     clean.Scope("metrics photos albums videos"),
		AuthProvider:  authn.ProviderAccessToken.String(),
		AuthMethod:    authn.MethodDefault.String(),
		GrantType:     authn.GrantPassword.String(),
		ClientName:    "alice_token_scope",
		user:          UserFixtures.Pointer("alice"),
		UserUID:       UserFixtures.Pointer("alice").UserUID,
		UserName:      UserFixtures.Pointer("alice").UserName,
		PreviewToken:  "cdd3r0lr",
		DownloadToken: "64ydcbom",
	},
	"bob": {
		authToken:   "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1",
		ID:          rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac1"),
		RefID:       "sessxkkcabce",
		SessTimeout: unix.Day * 3,
		SessExpires: unix.Now() + unix.Week,
		GrantType:   authn.GrantPassword.String(),
		user:        UserFixtures.Pointer("bob"),
		UserUID:     UserFixtures.Pointer("bob").UserUID,
		UserName:    UserFixtures.Pointer("bob").UserName,
	},
	"unauthorized": {
		authToken:   "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2",
		ID:          rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac2"),
		RefID:       "sessxkkcabcf",
		SessTimeout: unix.Day * 3,
		SessExpires: unix.Now() + unix.Week,
		GrantType:   authn.GrantImplicit.String(),
		user:        UserFixtures.Pointer("unauthorized"),
		UserUID:     UserFixtures.Pointer("unauthorized").UserUID,
		UserName:    UserFixtures.Pointer("unauthorized").UserName,
	},
	"visitor": {
		authToken:   "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3",
		ID:          rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3"),
		RefID:       "sessxkkcabcg",
		SessTimeout: unix.Day * 3,
		SessExpires: unix.Now() + unix.Week,
		user:        &Visitor,
		UserUID:     Visitor.UserUID,
		UserName:    Visitor.UserName,
		GrantType:   authn.GrantShareToken.String(),
		DataJSON:    []byte(`{"tokens":["1jxf3jfn2k"],"shares":["as6sg6bxpogaaba8"]}`),
		data: &SessionData{
			Tokens: []string{"1jxf3jfn2k"},
			Shares: UIDs{"as6sg6bxpogaaba8"},
		},
	},
	"visitor_token_metrics": {
		authToken:    "4ebe1048a7384e1e6af2930b5b6f29795ffab691df47a488",
		ID:           rnd.SessionID("4ebe1048a7384e1e6af2930b5b6f29795ffab691df47a488"),
		RefID:        "sessaae5cxun",
		SessTimeout:  0,
		SessExpires:  unix.Now() + unix.Week,
		AuthScope:    clean.Scope("metrics"),
		AuthProvider: authn.ProviderAccessToken.String(),
		AuthMethod:   authn.MethodDefault.String(),
		GrantType:    authn.GrantShareToken.String(),
		ClientName:   "visitor_token_metrics",
		user:         &Visitor,
		UserUID:      Visitor.UserUID,
		UserName:     Visitor.UserName,
	},
	"friend": {
		authToken:   "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac4",
		ID:          rnd.SessionID("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac4"),
		RefID:       "sessxkkcabch",
		SessTimeout: unix.Day * 3,
		SessExpires: unix.Now() + unix.Week,
		user:        UserFixtures.Pointer("friend"),
		UserUID:     UserFixtures.Pointer("friend").UserUID,
		UserName:    UserFixtures.Pointer("friend").UserName,
	},
	"client_metrics": {
		authToken:     "9d8b8801ffa23eb52e08ca7766283799ddfd8dd368212345",
		ID:            rnd.SessionID("9d8b8801ffa23eb52e08ca7766283799ddfd8dd368212345"),
		RefID:         "sessgh612345",
		SessTimeout:   0,
		SessExpires:   unix.Now() + unix.Week,
		AuthScope:     clean.Scope("metrics"),
		AuthProvider:  authn.ProviderClient.String(),
		AuthMethod:    authn.MethodOAuth2.String(),
		GrantType:     authn.GrantClientCredentials.String(),
		ClientUID:     ClientFixtures.Get("metrics").ClientUID,
		ClientName:    ClientFixtures.Get("metrics").ClientName,
		user:          nil,
		UserUID:       "",
		UserName:      "",
		PreviewToken:  "py212345",
		DownloadToken: "vgl12345",
	},
	"token_metrics": {
		authToken:     "9d8b8801ffa23eb52e08ca7766283799ddfd8dd368208a9b",
		ID:            rnd.SessionID("9d8b8801ffa23eb52e08ca7766283799ddfd8dd368208a9b"),
		RefID:         "sessgh6gjuo1",
		SessTimeout:   0,
		SessExpires:   unix.Now() + unix.Week,
		AuthScope:     clean.Scope("metrics"),
		AuthProvider:  authn.ProviderAccessToken.String(),
		AuthMethod:    authn.MethodDefault.String(),
		GrantType:     authn.GrantCLI.String(),
		ClientName:    "token_metrics",
		user:          nil,
		UserUID:       "",
		UserName:      "",
		PreviewToken:  "py2xrgr3",
		DownloadToken: "vgln2ffb",
	},
	"token_settings": {
		authToken:     "3f9684f7d3dd3d5b84edd43289c7fb5ca32ee73bd0233237",
		ID:            rnd.SessionID("3f9684f7d3dd3d5b84edd43289c7fb5ca32ee73bd0233237"),
		RefID:         "sessyugn54so",
		SessTimeout:   0,
		SessExpires:   unix.Now() + unix.Week,
		AuthScope:     clean.Scope("settings"),
		AuthProvider:  authn.ProviderAccessToken.String(),
		AuthMethod:    authn.MethodDefault.String(),
		GrantType:     authn.GrantCLI.String(),
		ClientName:    "token_settings",
		user:          nil,
		UserUID:       "",
		UserName:      "",
		PreviewToken:  "py2xrgr3",
		DownloadToken: "vgln2ffb",
	},
	"client_analytics": {
		authToken:     "9d8b8801ffa23eb52e08ca7766283799ddfd8dd368212123",
		ID:            rnd.SessionID("9d8b8801ffa23eb52e08ca7766283799ddfd8dd368212123"),
		RefID:         "sessgh6123yt",
		SessTimeout:   0,
		SessExpires:   unix.Now() + unix.Week,
		AuthScope:     clean.Scope("statistics"),
		AuthProvider:  authn.ProviderClient.String(),
		AuthMethod:    authn.MethodOAuth2.String(),
		GrantType:     authn.GrantCLI.String(),
		ClientUID:     ClientFixtures.Get("analytics").ClientUID,
		ClientName:    ClientFixtures.Get("analytics").ClientName,
		user:          nil,
		UserUID:       "",
		UserName:      "",
		PreviewToken:  "py2123yt",
		DownloadToken: "vgl123yt",
	},
	"client_invalid": {
		authToken:     "1831986451da7acf34690b703ff528f67bcf255e005270e9",
		ID:            rnd.SessionID("1831986451da7acf34690b703ff528f67bcf255e005270e9"),
		RefID:         "sessgh6123uy",
		SessTimeout:   0,
		SessExpires:   unix.Now() + unix.Week,
		AuthScope:     clean.Scope("undefined"),
		AuthProvider:  authn.ProviderClient.String(),
		AuthMethod:    authn.MethodUndefined.String(),
		GrantType:     authn.GrantCLI.String(),
		ClientUID:     ClientFixtures.Get("invalid_method").ClientUID,
		ClientName:    ClientFixtures.Get("invalid_method").ClientName,
		user:          nil,
		UserUID:       "",
		UserName:      "",
		PreviewToken:  "py2123uy",
		DownloadToken: "vgl123uy",
	},
}

// CreateSessionFixtures inserts known entities into the database for testing.
func CreateSessionFixtures() {
	for _, entity := range SessionFixtures {
		Db().Create(&entity)
	}
}
