package commands

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestAuthCommands(t *testing.T) {
	t.Run("List", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthCommands, []string{"auth", "ls"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "bob")
		assert.Contains(t, output, "visitor")
	})
	t.Run("ListAlice", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthCommands, []string{"auth", "ls", "alice"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.NotContains(t, output, "bob")
		assert.NotContains(t, output, "visitor")
	})
}

func TestAuthFindSession(t *testing.T) {
	fixture := entity.SessionFixtures.Get("alice_token")

	t.Run("Success", func(t *testing.T) {
		for _, id := range []string{fixture.AuthToken(), fixture.ID, fixture.RefID} {
			m, err := authFindSession(id)
			require.NoError(t, err)
			assert.Equal(t, fixture.RefID, m.RefID)
		}
	})
	t.Run("InvalidID", func(t *testing.T) {
		_, err := authFindSession("abc")
		assertExitCode(t, err, 2)
	})
	t.Run("NotFound", func(t *testing.T) {
		token := "0123456789abcdef0123456789abcdef0123456789abcdef"
		_, err := authFindSession(token)
		assertExitCode(t, err, 3)
		assert.NotContains(t, err.Error(), token)
	})
}

func TestAuthSessionLabel(t *testing.T) {
	created := time.Date(2026, 9, 20, 23, 30, 0, 0, time.UTC)

	t.Run("Client", func(t *testing.T) {
		m := &entity.Session{RefID: "sessgh6123yt", ClientName: "Analytics", AuthProvider: authn.ProviderClient.String(),
			GrantType: authn.GrantCLI.String(), CreatedAt: created}
		assert.Equal(t, "session sessgh6123yt (client 'Analytics', grant cli, created 2026-09-20)", authSessionLabel(m))
	})
	t.Run("App", func(t *testing.T) {
		m := &entity.Session{RefID: "sess6ey1ykya", ClientName: "backup", AuthProvider: authn.ProviderApplication.String(),
			UserName: "admin", GrantType: authn.GrantPassword.String(), CreatedAt: created, LastActive: created.Add(time.Hour).Unix()}
		assert.Equal(t, "session sess6ey1ykya (app 'backup', user 'admin', grant password, created 2026-09-20, last active 2026-09-21 00:30:00)", authSessionLabel(m))
	})
	t.Run("NoDetails", func(t *testing.T) {
		m := &entity.Session{RefID: "sessxkkcabcd", LastActive: -1}
		assert.Equal(t, "session sessxkkcabcd", authSessionLabel(m))
	})
	t.Run("QuotedNames", func(t *testing.T) {
		m := &entity.Session{RefID: "sess6ey1ykya", ClientName: "x', user 'admin\u009b", AuthProvider: authn.ProviderApplication.String(),
			UserName: "alice"}
		assert.Equal(t, "session sess6ey1ykya (app 'x'', user ''admin?', user 'alice')", authSessionLabel(m))
	})
	t.Run("NoSecrets", func(t *testing.T) {
		m := &entity.Session{RefID: "sessxkkcabcd", UserName: "alice"}
		m.SetAuthToken("69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac0")
		label := authSessionLabel(m)
		assert.NotContains(t, label, m.AuthToken())
		assert.NotContains(t, label, m.ID)
	})
}
