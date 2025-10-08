package avatar

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/service/http/safe"
)

var (
	// Stricter defaults for avatar images than the generic HTTP safe defaults.
	defaultTimeout       = 15 * time.Second
	defaultMaxSize int64 = 10 << 20 // 10 MiB for avatar images
)

// SafeDownload delegates avatar image downloads to the shared HTTP safe downloader
// with hardened defaults suitable for small image files.
// Callers may pass a partially filled safe.Options to override defaults.
func SafeDownload(destPath, rawURL string, opt *safe.Options) error {
	// Start with strict avatar defaults.
	o := &safe.Options{
		Timeout:      defaultTimeout,
		MaxSizeBytes: defaultMaxSize,
		AllowPrivate: false, // block private/loopback by default
		// Prefer images but allow others at low priority; MIME is validated later.
		Accept: "image/jpeg, image/png, */*;q=0.1",
	}
	if opt != nil {
		if opt.Timeout > 0 {
			o.Timeout = opt.Timeout
		}
		if opt.MaxSizeBytes > 0 {
			o.MaxSizeBytes = opt.MaxSizeBytes
		}
		// Bool has no sentinel; just copy the value.
		o.AllowPrivate = opt.AllowPrivate
		if strings.TrimSpace(opt.Accept) != "" {
			o.Accept = opt.Accept
		}
	}
	return safe.Download(destPath, rawURL, o)
}
