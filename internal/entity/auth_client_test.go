package entity

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt/report"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	m := NewClient()
	assert.Equal(t, "", m.AuthScope)
	assert.Equal(t, m.AuthScope, m.Scope())
	m.SetScope(" metrics WEBdav!")
	assert.Equal(t, "metrics webdav", m.AuthScope)
	assert.Equal(t, m.AuthScope, m.Scope())
}

func TestClient_GetData(t *testing.T) {
	c := NewClient()
	d := &ClientData{
		Theme: "portal",
	}

	c.SetData(d)

	first := c.GetData()
	second := c.GetData()

	assert.Same(t, first, second)
	assert.Equal(t, "portal", second.Theme)
}

func TestClient_GetData_OIDCFields(t *testing.T) {
	t.Run("RoundTripThroughDataJSON", func(t *testing.T) {
		src := NewClient()
		src.SetData(&ClientData{
			AllowGroups:     []string{"media-acme-admin", "media-acme-viewer"},
			AllowGroupRoles: map[string]string{"media-acme-admin": "admin", "media-acme-viewer": "viewer"},
			RedirectURIs:    []string{"https://photos.example.com/api/v1/oidc/redirect"},
		})

		// Read back from a fresh Client populated only via the marshaled JSON,
		// so the test exercises the json.Unmarshal path, not the m.data cache.
		dst := &Client{DataJSON: src.DataJSON}
		got := dst.GetData()

		assert.Equal(t, []string{"media-acme-admin", "media-acme-viewer"}, got.AllowGroups)
		assert.Equal(t, "admin", got.AllowGroupRoles["media-acme-admin"])
		assert.Equal(t, "viewer", got.AllowGroupRoles["media-acme-viewer"])
		assert.Equal(t, []string{"https://photos.example.com/api/v1/oidc/redirect"}, got.RedirectURIs)
	})

	t.Run("OmitEmpty", func(t *testing.T) {
		c := NewClient()
		c.SetData(&ClientData{Theme: "portal"})

		// Older rows without the OIDC fields must round-trip cleanly and report zero values.
		assert.NotContains(t, string(c.DataJSON), "allowGroups")
		assert.NotContains(t, string(c.DataJSON), "allowGroupRoles")
		assert.NotContains(t, string(c.DataJSON), "redirectUris")

		got := (&Client{DataJSON: c.DataJSON}).GetData()
		assert.Nil(t, got.AllowGroups)
		assert.Nil(t, got.AllowGroupRoles)
		assert.Nil(t, got.RedirectURIs)
	})
}

func TestFindClientByUID(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		expected := ClientFixtures.Get("alice")

		m := FindClientByUID("cs5gfen1bgxz7s9i")

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, m.UserUID, UserFixtures.Get("alice").UserUID)
		assert.Equal(t, expected.ClientUID, m.GetUID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Bob", func(t *testing.T) {
		expected := ClientFixtures.Get("bob")

		m := FindClientByUID("cs5gfsvbd7ejzn8m")

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, m.UserUID, UserFixtures.Get("bob").UserUID)
		assert.Equal(t, expected.ClientUID, m.GetUID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Metrics", func(t *testing.T) {
		expected := ClientFixtures.Get("metrics")

		m := FindClientByUID("cs5cpu17n6gj2qo5")

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Empty(t, m.UserUID)
		assert.Equal(t, expected.ClientUID, m.GetUID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Invalid", func(t *testing.T) {
		m := FindClientByUID("123")
		assert.Nil(t, m)
	})
}

func TestClient_NoUID(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		c := Client{ClientUID: ""}
		assert.True(t, c.InvalidUID())
	})
	t.Run("False", func(t *testing.T) {
		c := Client{ClientUID: "cs5cpu17n6gj2hgt"}
		assert.False(t, c.InvalidUID())
	})
}

func TestClient_NoName(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		c := Client{ClientName: ""}
		assert.True(t, c.NoName())
	})
	t.Run("False", func(t *testing.T) {
		c := Client{ClientName: "Test Test"}
		assert.False(t, c.NoName())
	})
}

func TestClient_SetRole(t *testing.T) {
	c := Client{ClientName: "Test Test"}
	assert.False(t, c.HasRole("admin"))
	c.SetRole("admin")
	assert.True(t, c.HasRole("admin"))
}

func TestClient_SetDisplayName(t *testing.T) {
	t.Run("InstanceReportSets", func(t *testing.T) {
		c := Client{}
		c.SetDisplayName("Reported", SrcAuto)
		assert.Equal(t, "Reported", c.DisplayName)
		assert.Equal(t, SrcAuto, c.NameSrc)
	})
	t.Run("InstanceUpdatesOwnValue", func(t *testing.T) {
		c := Client{DisplayName: "Old", NameSrc: SrcAuto}
		c.SetDisplayName("New", SrcAuto)
		assert.Equal(t, "New", c.DisplayName)
	})
	t.Run("ManualPinsOverAuto", func(t *testing.T) {
		c := Client{DisplayName: "Reported", NameSrc: SrcAuto}
		c.SetDisplayName("Pinned", SrcManual)
		assert.Equal(t, "Pinned", c.DisplayName)
		assert.Equal(t, SrcManual, c.NameSrc)
	})
	t.Run("AutoCannotOverrideManual", func(t *testing.T) {
		c := Client{DisplayName: "Pinned", NameSrc: SrcManual}
		c.SetDisplayName("Reported", SrcAuto)
		assert.Equal(t, "Pinned", c.DisplayName)
		assert.Equal(t, SrcManual, c.NameSrc)
	})
	t.Run("ManualClearUnpins", func(t *testing.T) {
		c := Client{DisplayName: "Pinned", NameSrc: SrcManual}
		c.SetDisplayName("", SrcManual)
		assert.Equal(t, "", c.DisplayName)
		assert.Equal(t, SrcAuto, c.NameSrc)
	})
	t.Run("AutoEmptyIgnored", func(t *testing.T) {
		c := Client{DisplayName: "Reported", NameSrc: SrcAuto}
		c.SetDisplayName("", SrcAuto)
		assert.Equal(t, "Reported", c.DisplayName)
	})
}

func TestClient_User(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		alice := ClientFixtures.Get("alice")

		m := alice.User()

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, "alice", m.UserName)
		assert.Equal(t, "uqxetse3cy5eo9z2", m.UserUID)
		assert.Equal(t, "admin", m.UserRole)

	})
	t.Run("Metrics", func(t *testing.T) {
		metrics := ClientFixtures.Get("metrics")

		m := metrics.User()

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Empty(t, m.UserName)
		assert.Empty(t, m.UserUID)
		assert.Empty(t, m.UserRole)

	})
	t.Run("UnknownUser", func(t *testing.T) {
		c := Client{ClientName: "test",
			UserUID: "123",
		}

		m := c.User()

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Empty(t, m.UserName)
		assert.Empty(t, m.UserUID)
		assert.Empty(t, m.UserRole)
	})
	t.Run("Bob", func(t *testing.T) {
		c := Client{ClientName: "bob",
			UserUID: "uqxc08w3d0ej2283",
		}

		m := c.User()

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, "bob", m.UserName)
		assert.Equal(t, "uqxc08w3d0ej2283", m.UserUID)
		assert.Equal(t, "admin", m.UserRole)
	})
}

func TestClient_SetUser(t *testing.T) {
	t.Run("John", func(t *testing.T) {
		c := Client{ClientName: "test"}
		u := &User{UserUID: "uqxc08w3d0ej2111", UserName: "john"}

		assert.Empty(t, c.User().UserName)

		m := c.SetUser(u)

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.NotEmpty(t, c.User().UserName)
	})
}

func TestClient_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = Client{ClientName: "test"}
		if err := m.Create(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("AlreadyExists", func(t *testing.T) {
		var m = ClientFixtures.Get("alice")
		err := m.Create()
		assert.Error(t, err)
	})
}

func TestClient_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := FindClientByUID("cs5cpu17n6gj2aaa")
		assert.Nil(t, c)

		var m = Client{ClientName: "New Client", ClientUID: "cs5cpu17n6gj2aaa"}
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		c = FindClientByUID("cs5cpu17n6gj2aaa")

		if c == nil {
			t.Fatal("result must not be nil")
		}
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = Client{ClientName: "David", ClientUID: "cs5cpu17n6gj2bbb"}
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		err := m.Delete()

		assert.NoError(t, err)
	})
	t.Run("EmptyUID", func(t *testing.T) {
		var m = Client{ClientName: "No UUID"}

		err := m.Delete()

		assert.Error(t, err)
	})
}

func TestClient_Deleted(t *testing.T) {
	var ptr *Client
	deletedAt := time.Now().UTC()
	zero := time.Time{}
	assert.False(t, ClientFixtures.Pointer("alice").Deleted())
	assert.False(t, (&Client{DeletedAt: &zero}).Deleted())
	assert.True(t, (&Client{DeletedAt: &deletedAt}).Deleted())
	assert.True(t, ptr.Deleted())
}

func TestClient_Disabled(t *testing.T) {
	var ptr *Client
	deletedAt := time.Now().UTC()
	assert.False(t, ClientFixtures.Pointer("alice").Disabled())
	assert.True(t, (&Client{AuthEnabled: false}).Disabled())
	assert.True(t, (&Client{AuthEnabled: true, DeletedAt: &deletedAt}).Disabled())
	assert.True(t, ptr.Disabled())
}

func TestClient_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = Client{ClientName: "Peter", ClientUID: "cs5cpu17n6gj2ddd"}

		assert.Empty(t, m.AuthScope)

		err := m.Updates(Client{AuthScope: "metrics"})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "metrics", m.AuthScope)
	})
}

func TestClient_NewSecret(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = Client{ClientName: "Anna", ClientUID: "cs5cpu17n6gj2eee"}
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		s, err := m.NewSecret()

		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, m.VerifySecret(s))
		assert.NotEmpty(t, s)
	})
	t.Run("EmptyUID", func(t *testing.T) {
		var m = Client{ClientName: "No UUID"}

		s, err := m.NewSecret()

		assert.Error(t, err)
		assert.False(t, m.VerifySecret(s))
		assert.Empty(t, s)
	})
}

func TestClient_SetSecret(t *testing.T) {
	t.Run("EmptyUID", func(t *testing.T) {
		var m = Client{ClientName: "No UUID"}

		err := m.SetSecret("123")

		assert.Error(t, err)
	})
	t.Run("InvalidSecret", func(t *testing.T) {
		var m = Client{ClientUID: "cs5cpu17n6gj2eee"}

		err := m.SetSecret("123")

		assert.Error(t, err)
	})
}

func TestClient_Provider(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		client := NewClient()
		assert.Equal(t, authn.ProviderClient, client.Provider())
	})
	t.Run("Alice", func(t *testing.T) {
		client := ClientFixtures.Get("alice")
		assert.Equal(t, authn.ProviderClient, client.Provider())
	})
	t.Run("Bob", func(t *testing.T) {
		client := ClientFixtures.Get("bob")
		assert.Equal(t, authn.ProviderClient, client.Provider())
	})
}

func TestClient_Method(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		client := NewClient()
		assert.Equal(t, authn.MethodOAuth2, client.Method())
	})
	t.Run("Alice", func(t *testing.T) {
		client := ClientFixtures.Get("alice")
		assert.Equal(t, authn.MethodOAuth2, client.Method())
	})
	t.Run("Bob", func(t *testing.T) {
		client := ClientFixtures.Get("bob")
		assert.Equal(t, authn.MethodOAuth2, client.Method())
	})
}

func TestClient_UpdateLastActive(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = Client{ClientName: "Anne", ClientUID: "cs5cpu17n6gj2fff"}
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, m.LastActive)

		c := m.UpdateLastActive(true)

		assert.NotEmpty(t, c.LastActive)
	})
	t.Run("EmptyUID", func(t *testing.T) {
		var m = Client{ClientName: "No UUID"}

		c := m.UpdateLastActive(true)

		assert.Empty(t, c.LastActive)
	})
}

func TestClient_Tokens(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		var m = Client{ClientName: "cs5cpu17n6gj2bbb", AuthTokens: 0}
		assert.Equal(t, int64(1), m.Tokens())
		m.SetTokens(0)
		assert.Equal(t, int64(1), m.Tokens())
		m.SetTokens(1)
		assert.Equal(t, int64(1), m.Tokens())
		m.SetTokens(10)
		assert.Equal(t, int64(10), m.Tokens())
	})
	t.Run("Unlimited", func(t *testing.T) {
		var m = Client{ClientName: "cs5cpu17n6gj2bbb", AuthTokens: -1}
		assert.Equal(t, int64(-1), m.Tokens())
	})
	t.Run("One", func(t *testing.T) {
		var m = Client{ClientName: "cs5cpu17n6gj2bbb", AuthTokens: 1}
		assert.Equal(t, int64(1), m.Tokens())
	})
	t.Run("Many", func(t *testing.T) {
		var m = Client{ClientName: "cs5cpu17n6gj2bbb", AuthTokens: 10}
		assert.Equal(t, int64(10), m.Tokens())
	})
}

func TestClient_EnforceAuthTokenLimit(t *testing.T) {
	t.Run("EmptyUID", func(t *testing.T) {
		var m = Client{ClientName: "No UUID"}

		r := m.EnforceAuthTokenLimit("")

		assert.Equal(t, r, 0)
	})
	t.Run("NoToken", func(t *testing.T) {
		var m = Client{ClientName: "David", ClientUID: "cs5cpu17n6gj2bbb"}

		r := m.EnforceAuthTokenLimit("")

		assert.Equal(t, r, 0)
	})
	t.Run("NegativeTokenLimit", func(t *testing.T) {
		var m = Client{ClientName: "David", ClientUID: "cs5cpu17n6gj2bbb", AuthTokens: -1}

		r := m.EnforceAuthTokenLimit("")

		assert.Equal(t, r, 0)
	})
	t.Run("KeepsReservedSession", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00010"
		client.AuthTokens = 2

		// The lowest ID is the one the tiebreak alone would delete first.
		ids := newClientSessions(t, client, 5, time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC))

		assert.Equal(t, 3, client.EnforceAuthTokenLimit(ids[0]))
		assertSessions(t, []string{ids[0], ids[4]}, ids[1:4])

		assert.Equal(t, 0, client.EnforceAuthTokenLimit(ids[0]))
		assertSessions(t, []string{ids[0], ids[4]}, nil)
	})
	t.Run("UnsetTokenLimitKeepsOne", func(t *testing.T) {
		client := NewClient()
		client.ClientUID = "cs5gfen1bgx00011"
		client.AuthTokens = 0

		ids := newClientSessions(t, client, 3, time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC))

		assert.Equal(t, 2, client.EnforceAuthTokenLimit(ids[0]))
		assertSessions(t, ids[:1], ids[1:])
	})
}

func TestClient_VerifySecret(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		expected := ClientFixtures.Get("alice")

		m := FindClientByUID("cs5gfen1bgxz7s9i")

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, expected.ClientUID, m.GetUID())
		assert.False(t, m.VerifySecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.VerifySecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.VerifySecret(""))
		assert.True(t, m.InvalidSecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.InvalidSecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.InvalidSecret(""))
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Metrics", func(t *testing.T) {
		expected := ClientFixtures.Get("metrics")

		m := FindClientByUID("cs5cpu17n6gj2qo5")

		if m == nil {
			t.Fatal("result must not be nil")
		}

		assert.Equal(t, expected.ClientUID, m.GetUID())
		assert.True(t, m.VerifySecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.VerifySecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.False(t, m.VerifySecret(""))
		assert.False(t, m.InvalidSecret("xcCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.InvalidSecret("aaCbOrw6I0vcoXzhnOmXhjpVSyFq0l0e"))
		assert.True(t, m.InvalidSecret(""))
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
}

func TestClient_Expires(t *testing.T) {
	t.Run("Metrics", func(t *testing.T) {
		m := ClientFixtures.Get("metrics")

		r := m.Expires()

		assert.Equal(t, r.String(), "1h0m0s")
	})
	t.Run("Alice", func(t *testing.T) {
		m := ClientFixtures.Get("alice")

		r := m.Expires()

		assert.Equal(t, r.String(), "24h0m0s")
	})
}

func TestClient_String(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		m := &Client{}
		assert.Equal(t, m.String(), "n/a")
	})
	t.Run("NewClient", func(t *testing.T) {
		m := NewClient()
		assert.Equal(t, m.String(), "n/a")
	})
	t.Run("Metrics", func(t *testing.T) {
		m := ClientFixtures.Get("metrics")
		assert.Equal(t, m.String(), "cs5cpu17n6gj2qo5")
	})
	t.Run("Alice", func(t *testing.T) {
		m := ClientFixtures.Get("alice")
		assert.Equal(t, m.String(), "cs5gfen1bgxz7s9i")
	})
	t.Run("Name", func(t *testing.T) {
		m := NewClient()
		m.ClientName = "Foo Bar"
		assert.Equal(t, m.String(), "Foo Bar")
	})
}

func TestClient_UserInfo(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		assert.Equal(t, report.NotAssigned, NewClient().UserInfo())
	})
	t.Run("Alice", func(t *testing.T) {
		assert.Equal(t, "alice", ClientFixtures.Pointer("alice").UserInfo())
	})
	t.Run("Metrics", func(t *testing.T) {
		assert.Equal(t, report.NotAssigned, ClientFixtures.Pointer("metrics").UserInfo())
	})
}

func TestClient_AuthInfo(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		assert.Equal(t, "Client (OAuth2)", NewClient().AuthInfo())
	})
	t.Run("Alice", func(t *testing.T) {
		assert.Equal(t, "Client (OAuth2)", ClientFixtures.Pointer("alice").AuthInfo())
	})
	t.Run("Metrics", func(t *testing.T) {
		assert.Equal(t, "Client (OAuth2)", ClientFixtures.Pointer("metrics").AuthInfo())
	})
}

func TestClient_Report(t *testing.T) {
	t.Run("Metrics", func(t *testing.T) {
		m := ClientFixtures.Get("metrics")

		rows, _ := m.Report(true)
		assert.NotEmpty(t, rows)
	})
}

func TestClient_SetFormValues(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var m = Client{ClientName: "Neo", ClientUID: "cs5cpu17n6gj3aab"}

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		var values = form.Client{
			ClientName:   "New Name",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   authn.MethodOAuth2.String(),
			AuthScope:    "test",
			AuthExpires:  4000,
			AuthTokens:   3,
			AuthEnabled:  false,
		}

		c := m.SetFormValues(values)

		assert.Equal(t, "New Name", c.ClientName)
		assert.Equal(t, int64(4000), c.AuthExpires)
		assert.Equal(t, int64(3), c.AuthTokens)
		assert.Equal(t, false, c.AuthEnabled)
	})
	t.Run("Success2", func(t *testing.T) {
		var m = Client{ClientName: "Neo", ClientUID: "cs5cpu17n6gj3aab", AuthTokens: -4}

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		var values = form.Client{
			ClientName:   "Annika",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   authn.MethodOAuth2.String(),
			AuthScope:    "metrics",
			AuthExpires:  -4000,
			AuthTokens:   -5,
			AuthEnabled:  true,
		}

		c := m.SetFormValues(values)

		assert.Equal(t, "Annika", c.ClientName)
		assert.Equal(t, int64(3600), c.AuthExpires)
		assert.Equal(t, int64(-1), c.AuthTokens)
		assert.Equal(t, true, c.AuthEnabled)
	})
	t.Run("Success3", func(t *testing.T) {
		var m = Client{ClientName: "Neo", ClientUID: "cs5cpu17n6gj3aab"}

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		var values = form.Client{
			ClientName:   "Friend",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   authn.MethodOAuth2.String(),
			AuthScope:    "test",
			AuthExpires:  4000000,
			AuthTokens:   3000000000,
			AuthEnabled:  true,
			UserUID:      "uqxqg7i1kperxvu7",
		}

		c := m.SetFormValues(values)

		assert.Equal(t, "Friend", c.ClientName)
		assert.Equal(t, int64(2678400), c.AuthExpires)
		assert.Equal(t, int64(2147483647), c.AuthTokens)
		assert.Equal(t, true, c.AuthEnabled)
	})
	t.Run("UseDefaults", func(t *testing.T) {
		var m = Client{ClientName: "Default", ClientUID: "cs5cpu17n6gj7y5r"}

		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		var values = form.Client{
			ClientName:   "Friend",
			AuthProvider: "",
			AuthMethod:   "",
			AuthScope:    "",
			AuthExpires:  -1,
			AuthTokens:   3000000000,
			AuthEnabled:  true,
			UserUID:      "uqxqg7i1kperxvu7",
		}

		c := m.SetFormValues(values)

		assert.Equal(t, "Friend", c.ClientName)
		assert.Equal(t, int64(3600), c.AuthExpires)
		assert.Equal(t, "*", c.AuthScope)
		assert.Equal(t, "oauth2", c.AuthMethod)
		assert.Equal(t, "client", c.AuthProvider)

	})
}

func TestClient_SetFormValues_Role(t *testing.T) {
	t.Run("SetValidRoleFromForm", func(t *testing.T) {
		m := Client{ClientName: "RoleTest", ClientUID: "cs5cpu17n6gj9r01"}

		// Persist once to align with other tests, though not required for SetFormValues.
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		// Apply role via form.
		c := m.SetFormValues(form.Client{ClientRole: "portal"})

		assert.Equal(t, "portal", c.ClientRole)
		assert.True(t, c.HasRole(acl.RolePortal))
		assert.False(t, c.HasRole(acl.RoleClient))
	})
	t.Run("InvalidRoleFromFormDefaultsToClient", func(t *testing.T) {
		m := Client{ClientName: "InvalidRole", ClientUID: "cs5cpu17n6gj9r02"}
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		// Unknown role → default to client.
		c := m.SetFormValues(form.Client{ClientRole: "superuser"})

		assert.Equal(t, "client", c.ClientRole)
		assert.True(t, c.HasRole(acl.RoleClient))
	})
	t.Run("ChangeRoleFromClientToAdmin", func(t *testing.T) {
		m := NewClient()
		m.ClientName = "ChangeRole"
		m.ClientUID = "cs5cpu17n6gj9r03"

		// Default role is "client".
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}
		assert.True(t, m.HasRole(acl.RoleClient))

		// Change to admin via form.
		m.SetFormValues(form.Client{ClientRole: "admin"})
		assert.True(t, m.HasRole(acl.RoleAdmin))
		assert.Equal(t, "admin", m.ClientRole)
	})
}

func TestClient_SetFormValues_AuthEnabledToggle(t *testing.T) {
	// Start enabled; attempt to disable via form should NOT flip to false.
	m := NewClient()
	m.ClientName = "ToggleEnabled"
	m.ClientUID = "cs5cpu17n6gj9r04"
	m.AuthEnabled = true
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	m.SetFormValues(form.Client{AuthEnabled: false})
	assert.True(t, m.AuthEnabled, "SetFormValues should not disable when AuthEnabled=false")

	// Now explicitly enable from false → true
	m.AuthEnabled = false
	m.SetFormValues(form.Client{AuthEnabled: true})
	assert.True(t, m.AuthEnabled)
}

func TestClient_SetFormValues_SetUser(t *testing.T) {
	t.Run("ByUID", func(t *testing.T) {
		m := NewClient()
		m.ClientName = "SetUserByUID"
		m.ClientUID = "cs5cpu17n6gj9r05"
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		uid := UserFixtures.Pointer("friend").UserUID
		c := m.SetFormValues(form.Client{UserUID: uid})

		assert.Equal(t, uid, c.UserUID)
		assert.Equal(t, uid, c.User().UserUID)
	})
	t.Run("ByUserName", func(t *testing.T) {
		m := NewClient()
		m.ClientName = "SetUserByName"
		m.ClientUID = "cs5cpu17n6gj9r06"
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		c := m.SetFormValues(form.Client{UserName: "alice"})

		assert.Equal(t, UserFixtures.Pointer("alice").UserUID, c.UserUID)
		assert.Equal(t, "alice", c.UserName)
		assert.Equal(t, "alice", c.User().UserName)
	})
	t.Run("UnknownUserNoChange", func(t *testing.T) {
		// Seed with a known user, then attempt to change to an unknown one.
		m := NewClient()
		m.ClientName = "UnknownUserNoChange"
		m.ClientUID = "cs5cpu17n6gj9r07"
		m.SetUser(UserFixtures.Pointer("bob"))
		if err := m.Save(); err != nil {
			t.Fatal(err)
		}

		prevUID := m.UserUID
		c := m.SetFormValues(form.Client{UserUID: "u0000000000000xx", UserName: "nonexistent"})

		assert.Equal(t, prevUID, c.UserUID)
		assert.Equal(t, "bob", c.UserName)
	})
}

func TestClient_AclRole_Resolution(t *testing.T) {
	t.Run("EmptyIsNone", func(t *testing.T) {
		m := &Client{ClientRole: ""}
		assert.Equal(t, acl.RoleNone, m.AclRole())
	})
	t.Run("ClientIsClient", func(t *testing.T) {
		m := &Client{ClientRole: "client"}
		assert.Equal(t, acl.RoleClient, m.AclRole())
	})
	t.Run("DeletedIsNone", func(t *testing.T) {
		deletedAt := time.Now().UTC()
		m := &Client{ClientRole: "client", DeletedAt: &deletedAt}
		assert.Equal(t, acl.RoleNone, m.AclRole())
		assert.False(t, m.HasRole(acl.RoleClient))
	})
}

func TestClient_SetRole_AliasNoneAndCase(t *testing.T) {
	m := &Client{}
	m.SetRole("NoNe")
	assert.True(t, m.HasRole(acl.RoleNone))

	m.SetRole("")
	assert.True(t, m.HasRole(acl.RoleNone))
}

func TestClient_SetFormValues_DoesNotOverrideUID(t *testing.T) {
	m := NewClient()
	m.ClientName = "KeepUID"
	m.ClientUID = "cs5cpu17n6gj9r08"
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	// Attempt to override with a different id via form; should be ignored.
	c := m.SetFormValues(form.Client{ClientID: "cs5cpu17n6gj9zzz", ClientName: "KeepUID2"})
	assert.Equal(t, "cs5cpu17n6gj9r08", c.ClientUID)
	assert.Equal(t, "KeepUID2", c.ClientName)
}

func TestClient_Validate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := Client{
			ClientName:   "test",
			ClientType:   "test",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   "basic",
			AuthScope:    "all",
		}

		err := m.Validate()

		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ClientNameEmpty", func(t *testing.T) {
		m := Client{
			ClientName:   "",
			ClientType:   "test",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   "basic",
			AuthScope:    "all",
		}

		err := m.Validate()

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("ClientTypeEmpty", func(t *testing.T) {
		m := Client{
			ClientName:   "test",
			ClientType:   "",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   "basic",
			AuthScope:    "all",
		}

		err := m.Validate()

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("AuthMethodEmpty", func(t *testing.T) {
		m := Client{
			ClientName:   "test",
			ClientType:   "test",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   "",
			AuthScope:    "all",
		}

		err := m.Validate()

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("AuthScopeEmpty", func(t *testing.T) {
		m := Client{
			ClientName:   "test",
			ClientType:   "test",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   "basic",
			AuthScope:    "",
		}

		err := m.Validate()

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("Deleted", func(t *testing.T) {
		deletedAt := time.Now().UTC()
		m := Client{
			ClientName:   "test",
			ClientType:   "test",
			AuthProvider: authn.ProviderClient.String(),
			AuthMethod:   "basic",
			AuthScope:    "all",
			DeletedAt:    &deletedAt,
		}

		err := m.Validate()

		if err == nil {
			t.Fatal("error expected")
		}
	})
}
func TestFindClientByNodeUUID(t *testing.T) {
	t.Run("node", func(t *testing.T) {
		expected := ClientFixtures.Get("node")

		m := FindClientByNodeUUID(expected.NodeUUID)

		if m == nil {
			t.Fatal("result should not be nil")
		}

		assert.Equal(t, ClientFixtures.Get("node").ClientName, m.ClientName)
		assert.Equal(t, expected.ClientUID, m.GetUID())
		assert.NotEmpty(t, m.CreatedAt)
		assert.NotEmpty(t, m.UpdatedAt)
	})
	t.Run("Invalid", func(t *testing.T) {
		m := FindClientByNodeUUID("123")
		assert.Nil(t, m)
	})
	t.Run("Empty", func(t *testing.T) {
		m := FindClientByNodeUUID("")
		assert.Nil(t, m)
	})
	t.Run("Deleted", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		m := NewClient().SetName("pp-retired-lookup")
		m.NodeUUID = uuid

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		// Both lookups see the record while it is current.
		assert.NotNil(t, FindClientByNodeUUID(uuid))
		assert.Len(t, FindClientsByNodeUUID(uuid), 1)

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		// Nothing resolves the retired node anymore, so no caller acts on it.
		assert.Nil(t, FindClientByNodeUUID(uuid))

		// The record is still reported by identifier, so the UUID counts as taken.
		retired := FindClientsByNodeUUID(uuid)

		if assert.Len(t, retired, 1) {
			assert.True(t, retired[0].Deleted())
			assert.Equal(t, m.ClientUID, retired[0].ClientUID)
		}

		// The client UID stays resolvable as well, so an ownership check can see it.
		byUID := FindClientByUID(m.ClientUID)

		if assert.NotNil(t, byUID) {
			assert.True(t, byUID.Deleted())
		}
	})
}

func TestFindClient(t *testing.T) {
	t.Run("ClientUID", func(t *testing.T) {
		expected := ClientFixtures.Get("alice")
		m := FindClient(expected.ClientUID)

		if assert.NotNil(t, m) {
			assert.Equal(t, expected.ClientUID, m.ClientUID)
		}
	})
	t.Run("NodeUUID", func(t *testing.T) {
		expected := ClientFixtures.Get("node")
		m := FindClient(expected.NodeUUID)

		if assert.NotNil(t, m) {
			assert.Equal(t, expected.ClientUID, m.ClientUID)
		}
	})
	t.Run("DeletedNodeUUID", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		c := NewClient().SetName("pp-find-retired")
		c.NodeUUID = uuid

		if err := c.Create(); err != nil {
			t.Fatal(err)
		}

		if err := c.Delete(); err != nil {
			t.Fatal(err)
		}

		// Operator tooling must still reach a retired record to restore or purge it.
		m := FindClient(uuid)

		if assert.NotNil(t, m) {
			assert.Equal(t, c.ClientUID, m.ClientUID)
			assert.True(t, m.Deleted())
		}

		assert.NotNil(t, FindClient(c.ClientUID))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Nil(t, FindClient(rnd.UUIDv7()))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, FindClient(""))
	})
}

func TestClient_Restore(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		m := NewClient().SetName("pp-restore").SetScope("metrics")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		if deleted := FindClientByUID(m.ClientUID); assert.NotNil(t, deleted) {
			assert.True(t, deleted.Deleted())
		}

		if err := m.Restore(); err != nil {
			t.Fatal(err)
		}

		assert.False(t, m.Deleted())

		// A restored client authenticates and resolves again.
		restored := FindClientByUID(m.ClientUID)

		if assert.NotNil(t, restored) {
			assert.False(t, restored.Deleted())
			assert.False(t, restored.Disabled())
			assert.NoError(t, restored.Validate())
		}
	})
	t.Run("NotDeleted", func(t *testing.T) {
		m := NewClient().SetName("pp-restore-live")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		assert.NoError(t, m.Restore())
		assert.False(t, m.Deleted())
	})
	t.Run("EmptyUID", func(t *testing.T) {
		assert.Error(t, (&Client{}).Restore())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Client
		assert.Error(t, m.Restore())
	})
}

func TestClient_Purge(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		m := NewClient().SetName("pp-purge")
		m.NodeUUID = uuid

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		// While retired, the UUID is still taken.
		assert.Len(t, FindClientsByNodeUUID(uuid), 1)

		if err := m.Purge(); err != nil {
			t.Fatal(err)
		}

		// Purging releases both identifiers.
		assert.Nil(t, FindClientByUID(m.ClientUID))
		assert.Empty(t, FindClientsByNodeUUID(uuid))
	})
	t.Run("NotDeleted", func(t *testing.T) {
		m := NewClient().SetName("pp-purge-live")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		// A purge does not require a prior delete.
		assert.NoError(t, m.Purge())
		assert.Nil(t, FindClientByUID(m.ClientUID))
	})
	t.Run("EmptyUID", func(t *testing.T) {
		assert.Error(t, (&Client{}).Purge())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Client
		assert.Error(t, m.Purge())
	})
}

func TestClient_RestoreConflict(t *testing.T) {
	t.Run("NameTakenByReplacement", func(t *testing.T) {
		// Deletion releases the client name, so a replacement may take it. Restoring the
		// original would then leave two current holders and lookups resolve by recency,
		// so the restored record could take the name over from the live one.
		original := NewClient().SetName("pp-conflict").SetScope("metrics")

		if err := original.Create(); err != nil {
			t.Fatal(err)
		}

		if err := original.Delete(); err != nil {
			t.Fatal(err)
		}

		replacement := NewClient().SetName("pp-conflict").SetScope("metrics")

		if err := replacement.Create(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "client name", original.RestoreConflict())
		assert.Error(t, original.Restore())
		assert.True(t, original.Deleted())

		// The live replacement keeps the name.
		if err := replacement.Delete(); err != nil {
			t.Fatal(err)
		}

		// Once the replacement is gone the original can come back.
		assert.Empty(t, original.RestoreConflict())
		assert.NoError(t, original.Restore())
		assert.False(t, original.Deleted())
	})
	t.Run("NodeUUIDTakenByDuplicate", func(t *testing.T) {
		uuid := rnd.UUIDv7()

		retired := NewClient().SetName("pp-conflict-uuid-a")
		retired.NodeUUID = uuid

		if err := retired.Create(); err != nil {
			t.Fatal(err)
		}

		if err := retired.Delete(); err != nil {
			t.Fatal(err)
		}

		live := NewClient().SetName("pp-conflict-uuid-b")
		live.NodeUUID = uuid

		if err := live.Create(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "node uuid", retired.RestoreConflict())
		assert.Error(t, retired.Restore())

		// The live record still resolves for the UUID.
		if got := FindClientByNodeUUID(uuid); assert.NotNil(t, got) {
			assert.Equal(t, live.ClientUID, got.ClientUID)
		}
	})
	t.Run("NoConflict", func(t *testing.T) {
		m := NewClient().SetName("pp-conflict-none")

		if err := m.Create(); err != nil {
			t.Fatal(err)
		}

		if err := m.Delete(); err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, m.RestoreConflict())
		assert.NoError(t, m.Restore())
	})
	t.Run("NotDeleted", func(t *testing.T) {
		m := NewClient().SetName("pp-conflict-live")
		assert.Empty(t, m.RestoreConflict())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Client
		assert.Empty(t, m.RestoreConflict())
	})
	t.Run("StaleReceiver", func(t *testing.T) {
		// A record read before the deletion carries no mark. The check reads the stored
		// record, so it still sees the conflict a second holder created since.
		original := NewClient().SetName("pp-conflict-stale").SetScope("metrics")

		if err := original.Create(); err != nil {
			t.Fatal(err)
		}

		stale := FindClientByUID(original.ClientUID)

		if err := original.Delete(); err != nil {
			t.Fatal(err)
		}

		replacement := NewClient().SetName("pp-conflict-stale").SetScope("metrics")

		if err := replacement.Create(); err != nil {
			t.Fatal(err)
		}

		if assert.NotNil(t, stale) {
			assert.False(t, stale.Deleted(), "the copy in hand does not carry the mark")
			assert.Equal(t, "client name", stale.RestoreConflict())
			assert.Error(t, stale.Restore())
		}

		// The stored record stays retired, so the name keeps one holder.
		if retired := FindClientByUID(original.ClientUID); assert.NotNil(t, retired) {
			assert.True(t, retired.Deleted())
		}
	})
}

func TestClient_HasInactiveUser(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var m *Client
		assert.False(t, m.HasInactiveUser())
	})
	t.Run("NoUser", func(t *testing.T) {
		assert.False(t, ClientFixtures.Pointer("metrics").HasInactiveUser())
	})
	t.Run("ActiveUser", func(t *testing.T) {
		assert.False(t, ClientFixtures.Pointer("alice").HasInactiveUser())
	})
	t.Run("MissingUser", func(t *testing.T) {
		m := NewClient()
		m.UserUID = rnd.GenerateUID(UserUID)
		assert.True(t, m.HasInactiveUser())
	})
	t.Run("InvalidUserUID", func(t *testing.T) {
		m := NewClient()
		m.UserUID = "invalid"
		assert.True(t, m.HasInactiveUser())
	})
	t.Run("DeletedUser", func(t *testing.T) {
		deletedAt := time.Now().Add(-time.Minute)
		u := *UserFixtures.Pointer("bob")
		u.DeletedAt = &deletedAt
		m := NewClient().SetUser(&u)
		assert.True(t, m.HasInactiveUser())
	})
	t.Run("ExpiredUser", func(t *testing.T) {
		expiresAt := time.Now().Add(-time.Minute)
		u := *UserFixtures.Pointer("bob")
		u.ExpiresAt = &expiresAt
		m := NewClient().SetUser(&u)
		assert.True(t, m.HasInactiveUser())
	})
	t.Run("ProviderNone", func(t *testing.T) {
		u := *UserFixtures.Pointer("bob")
		u.AuthProvider = authn.ProviderNone.String()
		m := NewClient().SetUser(&u)
		assert.True(t, m.HasInactiveUser())
	})
	t.Run("WebLoginDisabled", func(t *testing.T) {
		u := *UserFixtures.Pointer("bob")
		u.CanLogin = false
		m := NewClient().SetUser(&u)
		assert.False(t, m.HasInactiveUser())
	})
	t.Run("ExpiredSuperAdmin", func(t *testing.T) {
		expiresAt := time.Now().Add(-time.Minute)
		u := *UserFixtures.Pointer("alice")
		u.ExpiresAt = &expiresAt
		u.SuperAdmin = true
		m := NewClient().SetUser(&u)
		assert.False(t, m.HasInactiveUser())
	})
}
