package dl

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const minYtDlpVersion = "2025.09.23"

var (
	versionOnce    sync.Once
	versionWarning string
)

// VersionWarning returns a warning message when the detected yt-dlp version is older
// than the minimum recommended release. The check runs at most once per process.
func VersionWarning() (string, bool) {
	versionOnce.Do(func() {
		if os.Getenv("YTDLP_FAKE") == "1" {
			return
		}

		bin := FindYtDlpBin()
		if bin == "" {
			return
		}

		cmd := exec.Command(bin, "--version")
		cmd.Env = os.Environ()
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out

		if err := cmd.Run(); err != nil {
			return
		}

		first := strings.Fields(out.String())
		if len(first) == 0 {
			return
		}

		ver := first[0]
		if olderThan(ver, minYtDlpVersion) {
			versionWarning = "Detected yt-dlp " + ver + ". Please update to " + minYtDlpVersion + " or newer so YouTube videos expose playable formats."
		}
	})

	if versionWarning == "" {
		return "", false
	}
	return versionWarning, true
}

func olderThan(current, minimum string) bool {
	if current == "" {
		return false
	}

	c := truncateVersion(current)
	m := truncateVersion(minimum)

	if len(c) != len("2006.01.02") || len(m) != len("2006.01.02") {
		return false
	}

	ct, err1 := time.Parse("2006.01.02", c)
	mt, err2 := time.Parse("2006.01.02", m)
	if err1 != nil || err2 != nil {
		// fall back to lexical comparison which works for yyyy.mm.dd
		return c < m
	}
	return ct.Before(mt)
}

func truncateVersion(v string) string {
	if len(v) >= len("2006.01.02") {
		return v[:len("2006.01.02")]
	}
	return v
}

// ResetVersionWarningForTest resets the cached version warning; intended for tests only.
func ResetVersionWarningForTest() {
	versionOnce = sync.Once{}
	versionWarning = ""
}
