package dl

import (
	"bufio"
	"context"
	"os"
	"os/exec"
)

// ytDlpCommand builds the exec.Cmd for invoking yt-dlp.
// If the configured binary looks like a shell script (shebang),
// we invoke it via a shell to work around noexec mounts in CI.
func ytDlpCommand(ctx context.Context, args []string) *exec.Cmd {
	bin := FindYtDlpBin()

	// Optional override to force shell invocation.
	force := os.Getenv("YTDLP_FORCE_SHELL") == "1"

	if force || isShellScript(bin) {
		sh := os.Getenv("YTDLP_SHELL")
		if sh == "" {
			sh = "bash"
		}
		return exec.CommandContext(ctx, sh, append([]string{bin}, args...)...)
	}

	return exec.CommandContext(ctx, bin, args...)
}

// isShellScript tries to detect if a file starts with a shebang (#!).
func isShellScript(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	r := bufio.NewReader(f)
	b, err := r.Peek(2)
	if err != nil {
		return false
	}
	return len(b) >= 2 && b[0] == '#' && b[1] == '!'
}
