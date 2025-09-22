package dl

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// DownloadToFileWithOptions downloads media using yt-dlp into files on disk (no piping).
// It returns the list of files produced by yt-dlp, as printed via --print after_move:filepath.
func (result Metadata) DownloadToFileWithOptions(
	ctx context.Context,
	options DownloadOptions,
) ([]string, error) {
	if !result.Options.noInfoDownload {
		if (result.Info.Type == "playlist" ||
			result.Info.Type == "multi_video" ||
			result.Info.Type == "channel") &&
			options.PlaylistIndex == 0 {
			return nil, fmt.Errorf(
				"can't download a playlist when the playlist index options is not set",
			)
		}
	}

	tempPath, tempErr := os.MkdirTemp("", "ydls")
	if tempErr != nil {
		return nil, tempErr
	}
	defer os.RemoveAll(tempPath)

	var jsonTempPath string
	if !result.Options.noInfoDownload {
		jsonTempPath = filepath.Join(tempPath, "info.json")
		if err := os.WriteFile(jsonTempPath, result.RawJSON, 0600); err != nil {
			os.RemoveAll(tempPath)
			return nil, err
		}
	}

	cmd := exec.CommandContext(
		ctx,
		FindYtDlpBin(),
		// see comment below about ignoring errors for playlists
		"--ignore-errors",
		// TODO: deprecated in yt-dlp?
		"--no-call-home",
		// use non-fancy progress bar
		"--newline",
		// safer filenames
		"--restrict-filenames",
	)

	// Output template: caller may provide one; otherwise use a deterministic fallback in CWD
	// Note: caller should set a template rooted in the session temp dir.
	if options.Output != "" {
		cmd.Args = append(cmd.Args, "--output", options.Output)
	}

	// Print the final file paths after move/processing; also print plain filepath as a fallback
	cmd.Args = append(cmd.Args, "--print", "after_move:filepath")
	cmd.Args = append(cmd.Args, "--print", "filepath")

	if result.Options.noInfoDownload {
		cmd.Args = append(cmd.Args, "--batch-file", "-")
		cmd.Stdin = bytes.NewBufferString(result.RawURL + "\n")

		if result.Options.Type == TypePlaylist {
			cmd.Args = append(cmd.Args, "--yes-playlist")

			if result.Options.PlaylistStart > 0 {
				cmd.Args = append(cmd.Args,
					"--playlist-start", strconv.Itoa(int(result.Options.PlaylistStart)),
				)
			}
			if result.Options.PlaylistEnd > 0 {
				cmd.Args = append(cmd.Args,
					"--playlist-end", strconv.Itoa(int(result.Options.PlaylistEnd)),
				)
			}
			if result.Options.FlatPlaylist {
				cmd.Args = append(cmd.Args, "--flat-playlist")
			}
		} else {
			cmd.Args = append(cmd.Args,
				"--no-playlist",
			)
		}
	} else {
		cmd.Args = append(cmd.Args, "--load-info", jsonTempPath)
	}

	// force IPV4 Usage
	if result.Options.UseIPV4 {
		cmd.Args = append(cmd.Args, "-4")
	}
	// filter and playlist index
	if !result.Info.Direct && options.Filter != "" {
		cmd.Args = append(cmd.Args, "-f", options.Filter)
	}
	if options.PlaylistIndex > 0 {
		cmd.Args = append(cmd.Args, "--playlist-items", fmt.Sprint(options.PlaylistIndex))
	}
	if options.DownloadAudioOnly {
		cmd.Args = append(cmd.Args, "-x")
	}
	if options.EmbedMetadata {
		cmd.Args = append(cmd.Args, "--embed-metadata")
	}
	if options.EmbedSubs {
		cmd.Args = append(cmd.Args, "--embed-subs")
	}
	if options.ForceOverwrites {
		cmd.Args = append(cmd.Args, "--force-overwrites")
	}
	if options.DisableCaching {
		cmd.Args = append(cmd.Args, "--no-cache-dir")
	}
	if options.AudioFormats != "" {
		cmd.Args = append(cmd.Args, "--audio-format", options.AudioFormats)
	}
	if result.Options.ProxyUrl != "" {
		cmd.Args = append(cmd.Args, "--proxy", result.Options.ProxyUrl)
	}
	if result.Options.Downloader != "" {
		cmd.Args = append(cmd.Args, "--downloader", result.Options.Downloader)
	}
	if result.Options.DownloadSections != "" {
		cmd.Args = append(cmd.Args, "--download-sections", result.Options.DownloadSections)
	}
	if result.Options.CookiesFromBrowser != "" {
		cmd.Args = append(cmd.Args, "--cookies-from-browser", result.Options.CookiesFromBrowser)
	}
	if result.Options.Cookies != "" {
		cmd.Args = append(cmd.Args, "--cookies", result.Options.Cookies)
	}
	if result.Options.MergeOutputFormat != "" {
		cmd.Args = append(cmd.Args, "--merge-output-format", result.Options.MergeOutputFormat)
	}
	if result.Options.RemuxVideo != "" {
		cmd.Args = append(cmd.Args, "--remux-video", result.Options.RemuxVideo)
	}
	if result.Options.RecodeVideo != "" {
		cmd.Args = append(cmd.Args, "--recode-video", result.Options.RecodeVideo)
	}
	if result.Options.Fixup != "" {
		cmd.Args = append(cmd.Args, "--fixup", result.Options.Fixup)
	}
	if result.Options.SortingFormat != "" {
		cmd.Args = append(cmd.Args, "--format-sort", result.Options.SortingFormat)
	}
	if len(result.Options.AddHeaders) > 0 {
		for _, h := range result.Options.AddHeaders {
			if strings.TrimSpace(h) == "" {
				continue
			}
			cmd.Args = append(cmd.Args, "--add-header", h)
		}
	}

	cmd.Dir = tempPath

	if strings.TrimSpace(result.Options.FFmpegPostArgs) != "" {
		cmd.Args = append(cmd.Args, "--postprocessor-args", "ffmpeg:"+result.Options.FFmpegPostArgs)
	}

	// Capture stdout/stderr for parsing results and errors
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	log.Trace("cmd", " ", redactArgs(cmd.Args))
	err := cmd.Run()

	// Parse printed file paths from stdout
	var files []string
	scanner := bufio.NewScanner(bytes.NewReader(stdoutBuf.Bytes()))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// If relative, resolve against tempPath
		if !filepath.IsAbs(line) {
			line = filepath.Join(tempPath, line)
		}
		if _, statErr := os.Stat(line); statErr == nil {
			files = append(files, line)
		}
	}

	if err != nil {
		// Prefer returning the process error; callers can inspect stderr if needed
		return files, err
	}

	return files, nil
}
