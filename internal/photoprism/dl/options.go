package dl

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// Options for NewMetadata()
type Options struct {
	Type               Type
	PlaylistStart      uint   // --playlist-start
	PlaylistEnd        uint   // --playlist-end
	FlatPlaylist       bool   // --flat-playlist, faster fetching but with less video info for playlists
	Downloader         string // --downloader
	DownloadThumbnail  bool
	DownloadSubtitles  bool
	DownloadSections   string // --download-sections
	Impersonate        string // --impersonate
	ProxyUrl           string // --proxy URL  http://host:port or socks5://host:port
	UseIPV4            bool   // -4 Make all connections via IPv4
	Cookies            string // --cookies FILE
	CookiesFromBrowser string // --cookies-from-browser BROWSER[:FOLDER]
	// Note: The PhotoPrism CLI intentionally does NOT expose a --cookies-from-browser flag
	// because the default runtime is a container without access to local browser profiles.
	// This field remains for tests and non-container scenarios.
	AddHeaders        []string                      // --add-header "Name: Value" (repeatable)
	StderrFn          func(cmd *exec.Cmd) io.Writer // if not nil, function to get Writer for stderr
	HttpClient        *http.Client                  // Client for download thumbnail and subtitles (nil use http.DefaultClient)
	MergeOutputFormat string                        // --merge-output-format
	RemuxVideo        string                        // --remux-video
	RecodeVideo       string                        // --recode-video
	Fixup             string                        // --fixup
	SortingFormat     string                        // --format-sort
	// FFmpegPostArgs are additional ffmpeg post-processor args, e.g.,
	// "-metadata creation_time=2025-09-20T00:00:00Z". Applied via
	// yt-dlp's --postprocessor-args for the ffmpeg post-processor.
	FFmpegPostArgs string

	// Set to true if you don't want to use the result.Info structure after the goutubedl.NewMetadata() call,
	// so the given URL will be downloaded in a single pass in the DownloadResult.Download() call.
	noInfoDownload bool
}

type DownloadOptions struct {
	Filter            string // Download format matched by filter (usually a format id or quality designator).
	AudioFormats      string // --audio-formats Download audio using formats (best, aac, alac, flac, m4a, mp3, opus, vorbis, wav).
	DownloadAudioOnly bool   // -x Download audio only from video.
	EmbedMetadata     bool   // --embed-metadata embeds metadata to the video file.
	EmbedSubs         bool   // --embed-subs embeds subtitles in the video file
	ForceOverwrites   bool   // --force-overwrites replaces existing files
	DisableCaching    bool   // --no-cache-dir
	PlaylistIndex     int    // --playlist-items index of the file to download if there is more than one video
	Output            string
}

func (result Metadata) DownloadWithOptions(
	ctx context.Context,
	options DownloadOptions,
) (*DownloadResult, error) {
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

	var jsonTempPath string
	if !result.Options.noInfoDownload {
		jsonTempPath = path.Join(tempPath, "info.json")
		if err := os.WriteFile(jsonTempPath, result.RawJSON, 0600); err != nil {
			os.RemoveAll(tempPath)
			return nil, err
		}
	}

	dr := &DownloadResult{
		waitCh: make(chan struct{}),
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
		// use safer output filenmaes
		// TODO: needed?
		"--restrict-filenames",
		// use .netrc authentication data
		// "--netrc",
	)

	if options.Output != "" {
		cmd.Args = append(cmd.Args, "--output", options.Output)
	} else {
		cmd.Args = append(cmd.Args, "--output", "-")
	}

	if result.Options.noInfoDownload {
		// provide URL via stdin for security, youtube-dl has some run command args
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
	// don't need to specify if direct as there is only one
	// also seems to be issues when using filter with generic extractor
	if !result.Info.Direct && options.Filter != "" {
		cmd.Args = append(cmd.Args, "-f", options.Filter)
	}

	if options.PlaylistIndex > 0 {
		cmd.Args = append(cmd.Args, "--playlist-items", fmt.Sprint(options.PlaylistIndex))
	}

	if options.DownloadAudioOnly {
		cmd.Args = append(cmd.Args, "-x")
	}

	// If requested, embed metadata in the video file, including chapters and infoJSON,
	// see https://github.com/yt-dlp/yt-dlp?tab=readme-ov-file#post-processing-options.
	if options.EmbedMetadata {
		cmd.Args = append(cmd.Args, "--embed-metadata")
	}

	// If requested, embed subtitles in the video file.
	if options.EmbedSubs {
		cmd.Args = append(cmd.Args, "--embed-subs")
	}

	// If requested, overwrite existing video and metadata files.
	if options.ForceOverwrites {
		cmd.Args = append(cmd.Args, "--force-overwrites")
	}

	// If requested, disable filesystem caching.
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

	if len(result.Options.AddHeaders) > 0 {
		for _, h := range result.Options.AddHeaders {
			if strings.TrimSpace(h) == "" {
				continue
			}
			cmd.Args = append(cmd.Args, "--add-header", h)
		}
	}

	if result.Options.MergeOutputFormat != "" {
		cmd.Args = append(cmd.Args,
			"--merge-output-format", result.Options.MergeOutputFormat,
		)
	}

	if result.Options.RemuxVideo != "" {
		cmd.Args = append(cmd.Args,
			"--remux-video", result.Options.RemuxVideo,
		)
	}

	if result.Options.RecodeVideo != "" {
		cmd.Args = append(cmd.Args,
			"--recode-video", result.Options.RecodeVideo,
		)
	}

	if result.Options.Fixup != "" {
		cmd.Args = append(cmd.Args,
			"--fixup", result.Options.Fixup,
		)
	}

	if result.Options.SortingFormat != "" {
		cmd.Args = append(cmd.Args,
			"--format-sort", result.Options.SortingFormat,
		)
	}

	if strings.TrimSpace(result.Options.FFmpegPostArgs) != "" {
		cmd.Args = append(cmd.Args, "--postprocessor-args", "ffmpeg:"+result.Options.FFmpegPostArgs)
	}

	cmd.Dir = tempPath
	var stdoutW io.WriteCloser
	var stderrW io.WriteCloser
	var stderrR io.Reader
	dr.reader, stdoutW = io.Pipe()
	stderrR, stderrW = io.Pipe()
	optStderrWriter := io.Discard
	if result.Options.StderrFn != nil {
		optStderrWriter = result.Options.StderrFn(cmd)
	}
	cmd.Stdout = stdoutW
	cmd.Stderr = io.MultiWriter(optStderrWriter, stderrW)

	log.Trace("cmd", " ", redactArgs(cmd.Args))
	if err := cmd.Start(); err != nil {
		os.RemoveAll(tempPath)
		return nil, err
	}

	go func() {
		_ = cmd.Wait()
		stdoutW.Close()
		stderrW.Close()
		os.RemoveAll(tempPath)
		close(dr.waitCh)
	}()

	// blocks return until yt-dlp is downloading or has errored
	ytErrCh := make(chan error)
	go func() {
		stderrLineScanner := bufio.NewScanner(stderrR)
		for stderrLineScanner.Scan() {
			const downloadPrefix = "[download]"
			const errorPrefix = "ERROR: "
			line := stderrLineScanner.Text()
			if strings.HasPrefix(line, downloadPrefix) {
				break
			} else if strings.HasPrefix(line, errorPrefix) {
				ytErrCh <- errors.New(line[len(errorPrefix):])
				return
			}
		}
		ytErrCh <- nil
		_, _ = io.Copy(io.Discard, stderrR)
	}()

	return dr, <-ytErrCh
}
