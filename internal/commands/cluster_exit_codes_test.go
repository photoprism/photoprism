package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/service/cluster"
)

func TestExitCodes_Register_ValidationAndUnauthorized(t *testing.T) {
	t.Run("MissingURL", func(t *testing.T) {
		ctx := NewTestContext([]string{"register", "--name", "pp-node-01", "--role", "app", "--join-token", cluster.ExampleJoinToken})
		err := ClusterRegisterCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
}

func TestExitCodes_Nodes_PortalOnlyMisuse(t *testing.T) {
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
		ctx := NewTestContext([]string{"mod", "any", "--role", "app", "-y"})
		err := ClusterNodesModCommand.Action(ctx)
		assert.Error(t, err)
		if ec, ok := err.(cli.ExitCoder); ok {
			assert.Equal(t, 2, ec.ExitCode())
		} else {
			t.Fatalf("expected ExitCoder, got %T", err)
		}
	})
}
