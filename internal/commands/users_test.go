package commands

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/stretchr/testify/assert"
)

func TestUsersCommand(t *testing.T) {
	t.Run("AddModifyAndRemoveJohn", func(t *testing.T) {
		var err error

		//Add John
		args := []string{"add", "--name=John", "--email=john@test.de", "--password=test1234", "--role=admin", "john"}
		ctx := NewTestContext(args)

		// Run command with test context.
		output := capture.Output(func() {
			err = UsersAddCommand.Run(ctx, args...)
		})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		args2 := []string{"show", "john"}
		ctx2 := NewTestContext(args2)

		// Run command with test context.
		output2 := capture.Output(func() {
			err = UsersShowCommand.Run(ctx2, args2...)
		})

		//t.Logf(output2)
		assert.NoError(t, err)
		assert.Contains(t, output2, "John")
		assert.Contains(t, output2, "admin")
		assert.Contains(t, output2, "john@test.de")

		//Modify John

		// Create test context with flags and arguments.
		args3 := []string{"mod", "--name=Johnny", "--email=johnnny@test.de", "--password=test12345", "john"}
		ctx3 := NewTestContext(args3)

		// Run command with test context.
		output3 := capture.Output(func() {
			err = UsersModCommand.Run(ctx3, args3...)
		})

		// Check command output for plausibility.
		// t.Logf(output3)
		assert.NoError(t, err)
		assert.Empty(t, output3)

		output4 := capture.Output(func() {
			err = UsersShowCommand.Run(ctx2, args2...)
		})

		//t.Logf(output4)
		assert.NoError(t, err)
		assert.Contains(t, output4, "Johnny")
		assert.Contains(t, output4, "admin")
		assert.Contains(t, output4, "johnnny@test.de")
		assert.Contains(t, output4, "| DeletedAt     | <nil>")

		//Remove John
		// Create test context with flags and arguments.
		args5 := []string{"rm", "--force", "john"}
		ctx5 := NewTestContext(args5)

		// Run command with test context.
		output5 := capture.Output(func() {
			err = UsersRemoveCommand.Run(ctx5, args5...)
		})

		// Check command output for plausibility.
		// t.Logf(output5)
		assert.NoError(t, err)
		assert.Empty(t, output5)

		output6 := capture.Output(func() {
			err = UsersShowCommand.Run(ctx2, args2...)
		})

		//t.Logf(output6)
		assert.NoError(t, err)
		assert.Contains(t, output6, "Johnny")
		assert.Contains(t, output6, "admin")
		assert.Contains(t, output6, "johnnny@test.de")
		assert.Contains(t, output6, "| DeletedAt     | time.Date")
		assert.NotContains(t, output6, "| DeletedAt     | <nil>")
	})
}
