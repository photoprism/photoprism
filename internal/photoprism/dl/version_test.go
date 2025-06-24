package dl

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		versionRe := regexp.MustCompile(`^\d{4}\.\d{2}.\d{2}.*$`)
		version, versionErr := Version(context.Background())

		if versionErr != nil {
			t.Fatalf("err: %s", versionErr)
		}

		if !versionRe.MatchString(version) {
			t.Errorf("version %q does not match %q", version, versionRe)
		}
	})
	t.Run("InvalidBin", func(t *testing.T) {
		defer func(orig string) { YtDlpBin = orig }(YtDlpBin)
		YtDlpBin = "/non-existing"

		_, versionErr := Version(context.Background())
		if versionErr == nil || !strings.Contains(versionErr.Error(), "no such file or directory") {
			t.Fatalf("err should be nil 'no such file or directory': %v", versionErr)
		}
	})
}
