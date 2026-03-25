package commands

import (
	"strings"
	"testing"
)

// TestMCPCommandRegistered ensures the MCP command is present in the CLI catalog.
func TestMCPCommandRegistered(t *testing.T) {
	found := false

	for _, cmd := range PhotoPrism {
		if cmd.Name == "mcp" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected mcp command to be registered")
	}
}

// TestShowCommandsIncludesMCP ensures the command catalog exposes the MCP command.
func TestShowCommandsIncludesMCP(t *testing.T) {
	out, err := RunWithTestContext(ShowCommandsCommand, []string{"commands"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "## photoprism mcp") {
		t.Fatalf("expected MCP heading in output\n%s", out[:min(400, len(out))])
	}
}
