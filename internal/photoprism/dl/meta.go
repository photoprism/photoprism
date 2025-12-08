package dl

import (
	"math"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// CreatedFromInfo returns the best-effort creation time from Info.
// Priority:
// 1) Timestamp (UNIX seconds),
// 2) UploadDate (YYYYMMDD),
// 3) ReleaseDate (YYYYMMDD).
// Returns zero time if none can be parsed.
func CreatedFromInfo(info Info) time.Time {
	if info.Timestamp > 1 {
		sec, dec := math.Modf(info.Timestamp)
		return time.Unix(int64(sec), int64(dec*(1e9))).UTC()
	}
	parseYYYYMMDD := func(s string) (time.Time, bool) {
		s = strings.TrimSpace(s)
		if len(s) != 8 {
			return time.Time{}, false
		}
		t, err := time.Parse("20060102", s)
		if err != nil {
			return time.Time{}, false
		}
		return t.UTC(), true
	}
	if t, ok := parseYYYYMMDD(info.UploadDate); ok {
		return t
	}
	if t, ok := parseYYYYMMDD(info.ReleaseDate); ok {
		return t
	}
	return time.Time{}
}

// RemuxOptionsFromInfo builds ffmpeg remux options (container + metadata)
// based on yt-dlp Info and the source URL. The returned options enforce the
// target container and set title, description, author, comment, and created
// timestamp when provided by the extractor.
func RemuxOptionsFromInfo(ffmpegBin string, container fs.Type, info Info, sourceURL string) encode.Options {
	opt := encode.NewRemuxOptions(ffmpegBin, container, false)

	if title := clean.Name(info.Title); title != "" {
		opt.Title = title
	} else if title = clean.Name(info.AltTitle); title != "" {
		opt.Title = title
	}

	if desc := strings.TrimSpace(info.Description); desc != "" {
		opt.Description = desc
	}
	if u := strings.TrimSpace(sourceURL); u != "" {
		opt.Comment = u
	}

	if author := clean.Name(info.Artist); author != "" {
		opt.Author = author
	} else if author = clean.Name(info.AlbumArtist); author != "" {
		opt.Author = author
	} else if author = clean.Name(info.Creator); author != "" {
		opt.Author = author
	} else if author = clean.Name(info.License); author != "" {
		opt.Author = author
	}

	if created := CreatedFromInfo(info); !created.IsZero() {
		opt.Created = created
	}

	return opt
}
