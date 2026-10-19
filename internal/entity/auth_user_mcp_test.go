package entity

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestUser_CanUseMCP(t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		alice := UserFixtures.Get("alice")
		assert.Equal(t, 0, alice.MCP)
		assert.True(t, alice.CanUseMCP())
		alice.MCP = 1
		assert.True(t, alice.CanUseMCP())
		alice.MCP = -1
		assert.False(t, alice.CanUseMCP())
		alice.MCP = -2
		assert.False(t, alice.CanUseMCP())
	})
	t.Run("ProviderNone", func(t *testing.T) {
		alice := UserFixtures.Get("alice")
		alice.SetProvider(authn.ProviderNone)
		assert.False(t, alice.CanUseMCP())
	})
	t.Run("WebLoginDisabled", func(t *testing.T) {
		bob := UserFixtures.Get("bob")
		bob.CanLogin = false
		assert.True(t, bob.CanUseMCP())
	})
	t.Run("RoleWithoutMCP", func(t *testing.T) {
		guest := UserFixtures.Get("guest")
		guest.MCP = 1
		assert.False(t, guest.CanUseMCP())
		assert.False(t, UserFixtures.Pointer("jens.mander").CanUseMCP())
	})
	t.Run("Invalid", func(t *testing.T) {
		var m *User
		assert.False(t, m.CanUseMCP())
		assert.False(t, UserFixtures.Pointer("deleted").CanUseMCP())
		assert.False(t, (&User{UserRole: acl.RoleAdmin.String()}).CanUseMCP())
	})
}

func TestUser_MCPForm(t *testing.T) {
	t.Run("PrivilegeLevelChange", func(t *testing.T) {
		m := FindUserByName("alice")
		require.NotNil(t, m)
		frm, err := m.Form()
		require.NoError(t, err)
		assert.Equal(t, m.MCP, frm.MCP)
		assert.False(t, m.PrivilegeLevelChange(frm))
		frm.MCP = -1
		assert.True(t, m.PrivilegeLevelChange(frm))
		frm.MCP = 1
		assert.True(t, m.PrivilegeLevelChange(frm))
	})
	t.Run("SetFormValues", func(t *testing.T) {
		m := NewUser().SetFormValues(form.User{UserName: "mcp-form", UserRole: acl.RoleAdmin.String(), MCP: -5})
		assert.Equal(t, -1, m.MCP)
		m.SetFormValues(form.User{UserName: "mcp-form", UserRole: acl.RoleAdmin.String(), MCP: 3})
		assert.Equal(t, 1, m.MCP)
		m.SetFormValues(form.User{UserName: "mcp-form", UserRole: acl.RoleAdmin.String()})
		assert.Equal(t, 0, m.MCP)
	})
	t.Run("SaveForm", func(t *testing.T) {
		m := FirstOrCreateUser(&User{UserName: "saveformmcp", DisplayName: "Save Form MCP", UserRole: acl.RoleGuest.String()})
		require.NotNil(t, m)
		t.Cleanup(func() { deleteTestUser(t, m) })

		frm, err := m.Form()
		require.NoError(t, err)

		// Only admins change the setting.
		frm.MCP = -1
		require.NoError(t, m.SaveForm(frm, m, false, false))
		assert.Equal(t, 0, FindUserByUID(m.UserUID).MCP)

		require.NoError(t, m.SaveForm(frm, UserFixtures.Pointer("alice"), true, false))
		assert.Equal(t, -1, FindUserByUID(m.UserUID).MCP)

		// Setting it back stores the default.
		frm.MCP = 0
		require.NoError(t, m.SaveForm(frm, UserFixtures.Pointer("alice"), true, false))
		assert.Equal(t, 0, FindUserByUID(m.UserUID).MCP)

		frm.MCP = 9
		require.NoError(t, m.SaveForm(frm, UserFixtures.Pointer("alice"), true, false))
		assert.Equal(t, 1, FindUserByUID(m.UserUID).MCP)
	})
}

func TestUser_SetValuesFromCliMCP(t *testing.T) {
	newContext := func(t *testing.T, args ...string) *cli.Context {
		set := flag.NewFlagSet("users mod", flag.ContinueOnError)
		_ = set.Bool("mcp", false, "")
		require.NoError(t, set.Parse(args))
		return cli.NewContext(cli.NewApp(), set, nil)
	}

	m := FirstOrCreateUser(&User{UserName: "climcp", DisplayName: "CLI MCP", UserRole: acl.RoleAdmin.String()})
	require.NotNil(t, m)
	t.Cleanup(func() { deleteTestUser(t, m) })

	t.Run("False", func(t *testing.T) {
		// A change of the setting is a privilege-level change, which ends the account's login sessions.
		sess := NewSession(3600, 0)
		sess.SetUser(m)
		require.NoError(t, sess.Create())
		t.Cleanup(func() { _ = sess.Delete() })

		require.NoError(t, m.SetValuesFromCli(newContext(t, "--mcp=false")))
		assert.Equal(t, -1, m.MCP)
		assert.Error(t, UnscopedDb().First(&Session{}, "id = ?", sess.ID).Error)
	})
	t.Run("NotSet", func(t *testing.T) {
		require.NoError(t, m.SetValuesFromCli(newContext(t)))
		assert.Equal(t, -1, m.MCP)
	})
	t.Run("True", func(t *testing.T) {
		require.NoError(t, m.SetValuesFromCli(newContext(t, "--mcp")))
		assert.Equal(t, 1, m.MCP)
	})
}
