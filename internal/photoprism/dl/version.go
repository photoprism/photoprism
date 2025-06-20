package dl

import (
	"context"
	"os/exec"
	"strings"
)

// Version of youtube-dl.
// Might be a good idea to call at start to assert that youtube-dl can be found.
func Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, FindYtDlpBin(), "--version")
	versionBytes, cmdErr := cmd.Output()
	if cmdErr != nil {
		return "", cmdErr
	}

	return strings.TrimSpace(string(versionBytes)), nil
}
