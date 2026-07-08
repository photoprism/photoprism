package raw

import (
	"fmt"
	"os/exec"
	"strconv"
)

// DarktableOptions configures a Darktable RAW to JPEG conversion command.
type DarktableOptions struct {
	Bin       string // darktable-cli binary path
	RawName   string // source RAW file name
	XmpName   string // optional XMP sidecar file name
	JpegName  string // destination JPEG file name
	MaxSize   int    // maximum edge length in pixels
	Presets   bool   // apply custom presets (requires a global mutex)
	ConfigDir string // optional darktable config directory
	CacheDir  string // optional darktable cache directory
}

// DarktableCmd builds the Darktable CLI command and reports whether it requires a global mutex.
// Presets mode can run only one instance at a time, so the caller must serialize those invocations.
func DarktableCmd(o DarktableOptions) (*exec.Cmd, bool) {
	var args []string

	if o.XmpName != "" {
		args = []string{o.RawName, o.XmpName, o.JpegName}
	} else {
		args = []string{o.RawName, o.JpegName}
	}

	maxSize := strconv.Itoa(o.MaxSize)

	if o.Presets {
		args = append(args, "--width", maxSize, "--height", maxSize, "--hq", "true", "--upscale", "false")
	} else {
		// --apply-custom-presets=false disables locking.
		args = append(args, "--apply-custom-presets", "false", "--width", maxSize, "--height", maxSize, "--hq", "true", "--upscale", "false")
	}

	args = append(args, "--core", "--library", ":memory:")

	if o.ConfigDir != "" {
		args = append(args, "--configdir", o.ConfigDir)
	}

	if o.CacheDir != "" {
		args = append(args, "--cachedir", o.CacheDir)
	}

	// #nosec G204 -- arguments are built from validated config and file paths.
	return exec.Command(o.Bin, args...), o.Presets
}

// TherapeeCmd builds the RawTherapee CLI command that renders a RAW file to JPEG.
func TherapeeCmd(bin, rawName, jpegName, profile string, quality int) *exec.Cmd {
	jpegQuality := fmt.Sprintf("-j%d", quality)
	args := []string{"-o", jpegName, "-p", profile, "-s", "-d", jpegQuality, "-js3", "-b8", "-c", rawName}

	// #nosec G204 -- arguments are built from validated config and file paths.
	return exec.Command(bin, args...)
}

// ExifToolJpgFromRawCmd builds the ExifTool command that extracts the full-resolution embedded preview to stdout.
func ExifToolJpgFromRawCmd(bin, rawName string) *exec.Cmd {
	// #nosec G204 -- arguments are built from validated config and file paths.
	return exec.Command(bin, "-q", "-q", "-b", "-JpgFromRaw", rawName)
}

// ExifToolPreviewImageCmd builds the ExifTool command that extracts the smaller embedded preview to stdout.
func ExifToolPreviewImageCmd(bin, rawName string) *exec.Cmd {
	// #nosec G204 -- arguments are built from validated config and file paths.
	return exec.Command(bin, "-q", "-q", "-b", "-PreviewImage", rawName)
}
