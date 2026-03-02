package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// videoNormalizeFilter converts CLI args into a search query, mapping bare tokens to name/filename filters.
func videoNormalizeFilter(args []string) string {
	parts := make([]string, 0, len(args))

	for _, arg := range args {
		token := strings.TrimSpace(arg)
		if token == "" {
			continue
		}

		if strings.Contains(token, ":") {
			parts = append(parts, token)
			continue
		}

		if strings.Contains(token, "/") {
			parts = append(parts, fmt.Sprintf("filename:%s", token))
		} else {
			parts = append(parts, fmt.Sprintf("name:%s", token))
		}
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

// videoSplitTrimArgs separates filter args from the trailing trim duration argument.
func videoSplitTrimArgs(args []string) ([]string, string, error) {
	if len(args) == 0 {
		return nil, "", fmt.Errorf("missing duration argument")
	}

	filterArgs := make([]string, len(args)-1)
	copy(filterArgs, args[:len(args)-1])

	durationArg := strings.TrimSpace(args[len(args)-1])
	if durationArg == "" {
		return nil, "", fmt.Errorf("missing duration argument")
	}

	return filterArgs, durationArg, nil
}

// videoParseTrimDuration parses the trim duration string with the precedence and rules from the spec.
func videoParseTrimDuration(value string) (time.Duration, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return 0, fmt.Errorf("duration is empty")
	}

	sign := 1
	if strings.HasPrefix(raw, "-") {
		sign = -1
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "-"))
	}

	if raw == "" {
		return 0, fmt.Errorf("duration is empty")
	}

	if isDigits(raw) {
		secs, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q", value)
		}
		if secs == 0 {
			return 0, fmt.Errorf("duration must be non-zero")
		}
		return time.Duration(sign) * time.Duration(secs) * time.Second, nil
	}

	if strings.Contains(raw, ":") {
		if strings.ContainsAny(raw, "hms") {
			return 0, fmt.Errorf("invalid duration %q", value)
		}

		parts := strings.Split(raw, ":")
		if len(parts) != 2 && len(parts) != 3 {
			return 0, fmt.Errorf("invalid duration %q", value)
		}

		for _, p := range parts {
			if !isDigits(p) {
				return 0, fmt.Errorf("invalid duration %q", value)
			}
		}

		if len(parts) == 2 && len(parts[1]) != 2 {
			return 0, fmt.Errorf("invalid duration %q", value)
		}

		if len(parts) == 3 && (len(parts[1]) != 2 || len(parts[2]) != 2) {
			return 0, fmt.Errorf("invalid duration %q", value)
		}

		var hours, minutes, seconds int64

		if len(parts) == 2 {
			minutes, _ = strconv.ParseInt(parts[0], 10, 64)
			seconds, _ = strconv.ParseInt(parts[1], 10, 64)
		} else {
			hours, _ = strconv.ParseInt(parts[0], 10, 64)
			minutes, _ = strconv.ParseInt(parts[1], 10, 64)
			seconds, _ = strconv.ParseInt(parts[2], 10, 64)
		}

		total := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
		if total == 0 {
			return 0, fmt.Errorf("duration must be non-zero")
		}

		return time.Duration(sign) * total, nil
	}

	parsed, err := time.ParseDuration(applySign(raw, sign))
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q", value)
	}

	if parsed == 0 {
		return 0, fmt.Errorf("duration must be non-zero")
	}

	return parsed, nil
}

// videoListColumns returns the ordered column list for the video ls output.
func videoListColumns() []string {
	return []string{"Video", "Size", "Resolution", "Duration", "Frames", "FPS", "Content Type", "Checksum"}
}

// videoResultFiles returns the related files for a merged search result or falls back to the file fields on the result.
func videoResultFiles(found search.Photo) []entity.File {
	if len(found.Files) > 0 {
		return found.Files
	}

	return []entity.File{videoFileFromSearch(found)}
}

// videoFileFromSearch builds a file record from the file fields of a search result.
func videoFileFromSearch(found search.Photo) entity.File {
	return entity.File{
		ID:              found.FileID,
		PhotoUID:        found.PhotoUID,
		FileUID:         found.FileUID,
		FileRoot:        found.FileRoot,
		FileName:        found.FileName,
		OriginalName:    found.OriginalName,
		FileHash:        found.FileHash,
		FileWidth:       found.FileWidth,
		FileHeight:      found.FileHeight,
		FilePortrait:    found.FilePortrait,
		FilePrimary:     found.FilePrimary,
		FileSidecar:     found.FileSidecar,
		FileMissing:     found.FileMissing,
		FileVideo:       found.FileVideo,
		FileDuration:    found.FileDuration,
		FileFPS:         found.FileFPS,
		FileFrames:      found.FileFrames,
		FilePages:       found.FilePages,
		FileCodec:       found.FileCodec,
		FileType:        found.FileType,
		MediaType:       found.MediaType,
		FileMime:        found.FileMime,
		FileSize:        found.FileSize,
		FileOrientation: found.FileOrientation,
		FileProjection:  found.FileProjection,
		FileAspectRatio: found.FileAspectRatio,
		FileColors:      found.FileColors,
		FileDiff:        found.FileDiff,
		FileChroma:      found.FileChroma,
		FileLuminance:   found.FileLuminance,
		OmitMarkers:     true,
	}
}

// videoPrimaryFile selects the best video file from a merged search result, preferring non-sidecar entries.
func videoPrimaryFile(found search.Photo) (entity.File, bool) {
	files := videoResultFiles(found)
	if len(files) == 0 {
		return entity.File{}, false
	}

	for _, file := range files {
		if file.FileVideo && !file.FileSidecar {
			return file, true
		}
	}

	for _, file := range files {
		if file.FileVideo {
			return file, true
		}
	}

	return files[0], true
}

// videoListRow renders a search result row for table outputs with human-friendly values.
func videoListRow(found search.Photo) []string {
	videoFile, _ := videoPrimaryFile(found)

	row := []string{
		videoFile.FileName,
		videoHumanSize(videoFile.FileSize),
		fmt.Sprintf("%dx%d", videoFile.FileWidth, videoFile.FileHeight),
		videoHumanDuration(videoFile.FileDuration),
		videoHumanInt(videoFile.FileFrames),
		videoHumanFloat(videoFile.FileFPS),
		video.ContentType(videoFile.FileMime, videoFile.FileType, videoFile.FileCodec, videoFile.FileHDR),
		videoFile.FileHash,
	}

	return row
}

// videoListJSONRow renders a search result row for JSON output with canonical column keys.
func videoListJSONRow(found search.Photo) map[string]any {
	videoFile, _ := videoPrimaryFile(found)

	data := map[string]any{
		"video":        videoFile.FileName,
		"size":         videoNonNegativeSize(videoFile.FileSize),
		"resolution":   fmt.Sprintf("%dx%d", videoFile.FileWidth, videoFile.FileHeight),
		"duration":     videoFile.FileDuration.Nanoseconds(),
		"frames":       videoFile.FileFrames,
		"fps":          videoFile.FileFPS,
		"content_type": video.ContentType(videoFile.FileMime, videoFile.FileType, videoFile.FileCodec, videoFile.FileHDR),
		"checksum":     videoFile.FileHash,
	}

	return data
}

// videoListJSON marshals a list of JSON rows using the canonical keys for each column.
func videoListJSON(rows []map[string]any, cols []string) (string, error) {
	canon := make([]string, len(cols))
	for i, col := range cols {
		canon[i] = report.CanonKey(col)
	}

	payload := make([]map[string]any, 0, len(rows))

	for _, row := range rows {
		item := make(map[string]any, len(canon))
		for _, key := range canon {
			item[key] = row[key]
		}
		payload = append(payload, item)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// videoHumanDuration formats a duration for human-readable tables.
func videoHumanDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}

	return d.String()
}

// videoHumanInt formats non-zero integers for human-readable tables.
func videoHumanInt(value int) string {
	if value <= 0 {
		return ""
	}

	return strconv.Itoa(value)
}

// videoHumanFloat formats non-zero floats without unnecessary trailing zeros.
func videoHumanFloat(value float64) string {
	if value <= 0 {
		return ""
	}

	return strconv.FormatFloat(value, 'f', -1, 64)
}

// videoHumanSize formats file sizes with human-readable units.
func videoHumanSize(size int64) string {
	return humanize.Bytes(uint64(videoNonNegativeSize(size))) //nolint:gosec // size is bounded to non-negative values
}

// videoNonNegativeSize clamps negative sizes to zero before formatting.
func videoNonNegativeSize(size int64) int64 {
	if size < 0 {
		return 0
	}

	return size
}

// videoTempPath creates a temporary file path in the destination directory.
func videoTempPath(dir, pattern string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("temp directory is empty")
	}

	tmpFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}

	if err = tmpFile.Close(); err != nil {
		return "", err
	}

	if err = os.Remove(tmpFile.Name()); err != nil { //nolint:gosec // tmpFile path is generated by os.CreateTemp.
		return "", err
	}

	return tmpFile.Name(), nil
}

// videoFFmpegSeconds converts a duration into an ffmpeg-friendly seconds string.
func videoFFmpegSeconds(d time.Duration) string {
	seconds := d.Seconds()
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}

// isDigits reports whether the string contains only decimal digits.
func isDigits(value string) bool {
	if value == "" {
		return false
	}

	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// applySign applies a numeric sign to a duration string for parsing.
func applySign(value string, sign int) string {
	if sign >= 0 {
		return value
	}

	return "-" + value
}

// videoSidecarPath builds the sidecar destination path for an originals file without creating directories.
func videoSidecarPath(srcName, originalsPath, sidecarPath string) string {
	src := filepath.ToSlash(srcName)
	orig := filepath.ToSlash(originalsPath)

	if orig != "" {
		orig = strings.TrimSuffix(orig, "/") + "/"
	}

	rel := strings.TrimPrefix(src, orig)
	if rel == src {
		rel = filepath.Base(srcName)
	}

	rel = strings.TrimPrefix(rel, "/")
	return filepath.Join(sidecarPath, filepath.FromSlash(rel))
}
