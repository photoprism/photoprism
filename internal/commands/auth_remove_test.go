package commands

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// restoreAnalyticsSession puts the session fixture a subtest deletes back once the subtest ends.
func restoreAnalyticsSession(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		s := entity.SessionFixtures.Get("client_analytics")

		if err := reopenConnection().Db().FirstOrCreate(&s, "id = ?", s.ID).Error; err != nil {
			t.Errorf("restore session fixture: %s", err)
		}
	})
}

func TestAuthRemoveCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		output0, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Logf(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)
		assert.Contains(t, output0, "sessgh6123yt")

		t.Setenv("PHOTOPRISM_CLI", "")

		// Without a terminal the prompt cannot run, which is a usage error rather than a refusal.
		_, err = RunWithTestContext(AuthRemoveCommand, []string{"rm", "sessgh6123yt"})

		var exit cli.ExitCoder
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 2, exit.ExitCode())
		assert.Contains(t, err.Error(), "--yes")

		output1, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Logf(output1)
		assert.NoError(t, err)
		assert.NotEmpty(t, output1)
		assert.Contains(t, output1, "sessgh6123yt")
	})
	t.Run("NonInteractive", func(t *testing.T) {
		output0, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Log(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)
		assert.Contains(t, output0, "sessgh6123yt")

		restoreAnalyticsSession(t)
		t.Setenv(config.EnvVar("cli"), "noninteractive")

		// Setup and capture output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "sessgh6123yt"})
		// Reset logger
		log.SetOutput(os.Stdout)

		// t.Log(output)
		// t.Log(buffer.String())
		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Contains(t, buffer.String(), "session sessgh6123yt (client 'Analytics', grant cli, created ")
		assert.Contains(t, buffer.String(), ") has been removed")

		output1, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Log(output1)
		assert.Error(t, err)
		assert.Empty(t, output1)
		assertExitCode(t, err, 3)
		assert.Equal(t, "session not found", err.Error())
	})
	t.Run("Yes", func(t *testing.T) {
		restoreAnalyticsSession(t)
		t.Setenv("PHOTOPRISM_CLI", "")

		_, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "--yes", "sessgh6123yt"})
		require.NoError(t, err)

		_, err = RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})
		assert.Error(t, err)
	})
	t.Run("Identifiers", func(t *testing.T) {
		fixture := entity.SessionFixtures.Get("client_analytics")
		token := fixture.AuthToken()
		require.NotEmpty(t, token)

		// Each identifier form removes the session, which is named by its reference ID only.
		for name, id := range map[string]string{"AuthToken": token, "SessionID": fixture.ID, "RefID": fixture.RefID} {
			t.Run(name, func(t *testing.T) {
				restoreAnalyticsSession(t)
				t.Setenv("PHOTOPRISM_CLI", "")

				buffer := bytes.Buffer{}
				log.SetOutput(&buffer)
				t.Cleanup(func() { log.SetOutput(os.Stdout) })

				output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "--yes", id})
				require.NoError(t, err)

				all := output + buffer.String()
				assert.Contains(t, all, "session sessgh6123yt (client 'Analytics'")
				assert.NotContains(t, all, token)
				assert.NotContains(t, all, fixture.ID)
				assert.Nil(t, entity.FindSessionByRefID(fixture.RefID))
			})
		}
	})
	t.Run("Declined", func(t *testing.T) {
		fixture := entity.SessionFixtures.Get("client_analytics")
		token := fixture.AuthToken()
		t.Setenv("PHOTOPRISM_CLI", "")
		pipeResetAnswers(t, "n\n")

		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		t.Cleanup(func() { log.SetOutput(os.Stdout) })

		output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", token})
		require.NoError(t, err)

		// The prompt and the log line name the session, not the token it was looked up by.
		assert.Contains(t, output, "Remove session sessgh6123yt (client 'Analytics'")
		assert.Contains(t, buffer.String(), "session sessgh6123yt was not removed")
		assert.NotContains(t, output+buffer.String(), token)
		assert.NotContains(t, output+buffer.String(), fixture.ID)
		assert.NotNil(t, entity.FindSessionByRefID(fixture.RefID))
	})
	t.Run("NotFound", func(t *testing.T) {
		token := "0123456789abcdef0123456789abcdef0123456789abcdef"

		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		t.Cleanup(func() { log.SetOutput(os.Stdout) })

		output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "--yes", token})
		assertExitCode(t, err, 3)
		assert.Equal(t, "session not found", err.Error())
		assert.NotContains(t, output+buffer.String(), token)
	})
	t.Run("NoArgument", func(t *testing.T) {
		_, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "--yes"})
		assertExitCode(t, err, 2)
	})
	t.Run("NotFoundBeforePrompt", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_CLI", "")

		// The lookup comes first, so a missing session is reported without asking to confirm.
		_, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "sessxxxxxxxx"})
		assertExitCode(t, err, 3)
	})
	t.Run("InvalidID", func(t *testing.T) {
		_, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "--yes", "abc"})
		assertExitCode(t, err, 2)
		assert.Contains(t, err.Error(), "invalid session id")
	})
}

// TestAuthRemoveCommandDatabaseErrors checks runtime exits without removing the selected session.
func TestAuthRemoveCommandDatabaseErrors(t *testing.T) {
	for _, name := range []string{"Lookup", "Delete"} {
		t.Run(name, func(t *testing.T) {
			previousInit := InitConfig
			conf := reopenConnection()
			InitConfig = func(*cli.Context) (*config.Config, error) { return conf, nil }
			t.Cleanup(func() { conf.RegisterDb(); InitConfig = previousInit })
			session := entity.NewSession(3600, 0)
			require.NoError(t, session.Create())
			t.Cleanup(func() { conf.RegisterDb(); require.NoError(t, session.Delete()) })
			flags := flag.NewFlagSet("auth-rm-database", flag.ContinueOnError)
			for _, option := range AuthRemoveCommand.Flags {
				require.NoError(t, option.Apply(flags))
			}
			require.NoError(t, flags.Parse([]string{"--yes", session.RefID}))
			ctx := cli.NewContext(cli.NewApp(), flags, nil)
			processor, before := conf.Db().Callback().Query(), "gorm:query"
			if name == "Delete" {
				processor, before = conf.Db().Callback().Delete(), "gorm:delete"
			}
			callback := "test:auth-rm:" + name
			injected := errors.New("test session database error")
			calls := 0
			processor.Before(before).Register(callback, func(scope *gorm.Scope) {
				if scope.TableName() != (entity.Session{}).TableName() {
					return
				}
				if name == "Delete" {
					value, ok := scope.Value.(*entity.Session)
					if !ok || value.ID != session.ID {
						return
					}
				}
				calls++
				_ = scope.Err(injected)
			})
			runErr := func() error {
				defer processor.Remove(callback)
				return AuthRemoveCommand.Action(ctx)
			}()
			assertExitCode(t, runErr, 1)
			require.ErrorContains(t, runErr, injected.Error())
			assert.Equal(t, 1, calls)
			conf.RegisterDb()
			var stored entity.Session
			require.NoError(t, conf.Db().Where("id = ?", session.ID).First(&stored).Error)
			require.NoError(t, AuthRemoveCommand.Action(ctx))
			conf.RegisterDb()
			assert.True(t, gorm.IsRecordNotFoundError(conf.Db().Where("id = ?", session.ID).First(&stored).Error))
		})
	}
}
