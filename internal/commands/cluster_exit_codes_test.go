package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
)

// TestExitCodes_Register_ValidationAndUnauthorized checks registration errors and restores the database.
func TestExitCodes_Register_ValidationAndUnauthorized(t *testing.T) {
	previous := requireTestDb(t)
	t.Cleanup(func() {
		require.True(t, previous.IsDbOpen(), "direct command tests must restore the database")
		require.Same(t, previous.Db(), entity.Db())
	})
	t.Cleanup(previous.RegisterDb)
	t.Run("MissingURL", func(t *testing.T) {
		ctx := NewTestContext([]string{"register", "--name", "pp-node-01", "--role", "instance", "--join-token", cluster.ExampleJoinToken})
		err := ClusterRegisterCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
}

// TestExitCodes_Nodes_PortalOnlyMisuse checks role errors and restores the database.
func TestExitCodes_Nodes_PortalOnlyMisuse(t *testing.T) {
	previous := requireTestDb(t)
	t.Cleanup(func() {
		require.True(t, previous.IsDbOpen(), "direct command tests must restore the database")
		require.Same(t, previous.Db(), entity.Db())
	})
	t.Cleanup(previous.RegisterDb)
	t.Run("ListNotPortal", func(t *testing.T) {
		ctx := NewTestContext([]string{"ls"})
		err := ClusterNodesListCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
	t.Run("ShowNotPortal", func(t *testing.T) {
		ctx := NewTestContext([]string{"show", "any"})
		err := ClusterNodesShowCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
	t.Run("RemoveNotPortal", func(t *testing.T) {
		ctx := NewTestContext([]string{"rm", "any"})
		err := ClusterNodesRemoveCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
	t.Run("ModNotPortal", func(t *testing.T) {
		ctx := NewTestContext([]string{"mod", "any", "--role", "instance", "-y"})
		err := ClusterNodesModCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
}
