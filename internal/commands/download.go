package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/dl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var downloadExamples = `
Usage examples:

photoprism dl --cookies cookies.txt \
 --header 'Authorization: Bearer <token>' \
 --method file --remux auto -- \
 https://example.com/a.mp4 https://example.com/b.jpg

photoprism dl -a 'Authorization: Bearer <token>' \
			 -a 'Accept: application/json' -- URL`

// DownloadCommand configures the command name, flags, and action.
var DownloadCommand = &cli.Command{
	Name:        "download",
	Aliases:     []string{"dl"},
	Usage:       "Imports media from one or more URLs",
	Description: "Download and import media from one or more URLs.\n" + downloadExamples,
	ArgsUsage:   "[url]...",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "dest",
			Aliases: []string{"d"},
			Usage:   "relative originals `PATH` in which new files should be imported",
		},
		&cli.StringFlag{
			Name:    "impersonate",
			Aliases: []string{"i"},
			Usage:   "impersonate browser `IDENTITY` (e.g. chrome, edge or safari; 'none' to disable)",
			Value:   "firefox",
		},
		&cli.StringFlag{
			Name:    "method",
			Aliases: []string{"m"},
			Value:   "pipe",
			Usage:   "download `METHOD` when using external commands: pipe (stdio stream) or file (temporary files)",
		},
		&cli.StringFlag{
			Name:    "remux",
			Aliases: []string{"r"},
			Value:   "auto",
			Usage:   "remux `POLICY` for videos when using --method file: auto (skip if MP4), always, or skip",
		},
		&cli.StringFlag{
			Name:    "sort",
			Aliases: []string{"s"},
			Usage:   "custom `FORMAT` sort expression, e.g. 'quality,res,fps,codec:avc:m4a,size,br,asr,proto,ext,hasaud,source,id'",
		},
		&cli.StringFlag{
			Name:    "cookies",
			Aliases: []string{"c"},
			Usage:   "use Netscape-format cookies.txt `FILE` for HTTP authentication",
		},
		&cli.StringSliceFlag{
			Name:    "header",
			Aliases: []string{"a"},
			Usage:   "add HTTP request `HEADER` in the form 'Name: Value' (repeatable)",
		},
	},
	Action: downloadAction,
}

// downloadAction downloads and import media from a URL.
func downloadAction(ctx *cli.Context) error {
	start := time.Now()

	conf, confErr := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if confErr != nil {
		return confErr
	}

	// very if copy directory exist and is writable
	if conf.ReadOnly() {
		return config.ErrReadOnly
	}

	conf.InitDb()
	defer conf.Shutdown()

	// Collect URLs: args or STDIN when no args
	var inputURLs []string
	if ctx.Args().Len() > 0 {
		inputURLs = append(inputURLs, ctx.Args().Slice()...)
	} else {
		// If STDIN is a pipe, read URLs line by line (Phase 1: args take precedence; no --stdin merge)
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				inputURLs = append(inputURLs, line)
			}
			if err := scanner.Err(); err != nil {
				return err
			}
		}
	}

	if len(inputURLs) == 0 {
		return fmt.Errorf("no download URLs provided")
	}

	var destFolder string
	if ctx.IsSet("dest") {
		destFolder = clean.UserPath(ctx.String("dest"))
	} else {
		destFolder = conf.ImportDest()
	}

	var downloadPath, downloadFile string

	downloadPath = filepath.Join(conf.TempPath(), fs.DownloadDir+"_"+rnd.Base36(12))

	if err := fs.MkdirAll(downloadPath); err != nil {
		return err
	}

	defer os.RemoveAll(downloadPath)

	// Flags for yt-dlp auth and headers
	cookies := strings.TrimSpace(ctx.String("cookies"))

	// cookiesFromBrowser := strings.TrimSpace(ctx.String("cookies-from-browser"))
	addHeaders := ctx.StringSlice("header")

	impersonate := strings.ToLower(strings.TrimSpace(ctx.String("impersonate")))

	switch impersonate {
	case "":
		impersonate = "firefox"
	case "none":
		impersonate = ""
	}

	flagMethod := ""

	if ctx.IsSet("method") {
		flagMethod = ctx.String("method")
	}

	method, _, err := resolveDownloadMethod(flagMethod)

	if err != nil {
		return err
	}

	formatSort := strings.TrimSpace(ctx.String("sort"))
	sortingFormat := formatSort

	if sortingFormat == "" && method == "pipe" {
		sortingFormat = pipeSortingFormat
	}

	fileRemux := strings.ToLower(strings.TrimSpace(ctx.String("remux")))

	if fileRemux == "" {
		fileRemux = "auto"
	}

	switch fileRemux {
	case "always", "auto", "skip":
	default:
		return fmt.Errorf("invalid --remux: %s (expected 'always', 'auto', or 'skip')", fileRemux)
	}

	// Process inputs sequentially (Phase 1)
	var failures int
	for _, raw := range inputURLs {
		u, perr := url.Parse(strings.TrimSpace(raw))
		if perr != nil {
			log.Errorf("invalid URL: %s", clean.Log(raw))
			failures++
			continue
		}
		if u.Scheme != scheme.Http && u.Scheme != scheme.Https {
			log.Errorf("invalid URL scheme %s: %s", clean.Log(u.Scheme), clean.Log(raw))
			failures++
			continue
		}

		mt := media.FromName(u.Path)
		ext := fs.Ext(u.Path)

		switch mt {
		case media.Image, media.Vector, media.Raw, media.Document, media.Audio:
			log.Infof("downloading %s from %s", mt, clean.Log(u.String()))
			if dlName := clean.DlName(fs.BasePrefix(u.Path, true)); dlName != "" {
				downloadFile = dlName + ext
			} else {
				downloadFile = time.Now().Format("20060102_150405") + ext
			}
			downloadFilePath := filepath.Join(downloadPath, downloadFile)
			if downloadErr := fs.Download(downloadFilePath, u.String()); downloadErr != nil {
				log.Errorf("download failed: %v", downloadErr)
				failures++
				continue
			}
		default:
			mt = media.Video
			log.Infof("downloading %s from %s", mt, clean.Log(u.String()))

			opt := dl.Options{
				SortingFormat: sortingFormat,
				Cookies:       cookies,
				AddHeaders:    addHeaders,
				Impersonate:   impersonate,
			}
			ytRemux := method != "pipe"
			if ytRemux {
				opt.MergeOutputFormat = fs.VideoMp4.String()
				opt.RemuxVideo = fs.VideoMp4.String()
			}

			result, metaErr := dl.NewMetadata(context.Background(), u.String(), opt)

			if metaErr != nil {
				log.Errorf("metadata failed: %v", metaErr)
				if hint, ok := missingFormatsHint(metaErr); ok {
					log.Info(hint)
				}
				failures++
				continue
			}

			// Best-effort creation time for file method when not remuxing locally.
			if ytRemux {
				if created := dl.CreatedFromInfo(result.Info); !created.IsZero() {
					// Apply via yt-dlp ffmpeg post-processor so creation_time exists even without our remux.
					result.Options.FFmpegPostArgs = "-metadata creation_time=" + created.UTC().Format(time.RFC3339)
				}
			}

			// Base filename for pipe method
			if dlName := clean.DlName(result.Info.Title); dlName != "" {
				downloadFile = dlName + fs.ExtMp4
			} else {
				downloadFile = time.Now().Format("20060102_150405") + fs.ExtMp4
			}
			downloadFilePath := filepath.Join(downloadPath, downloadFile)

			if method == "pipe" {
				// Stream to stdout
				downloadResult, dlErr := dl.Download(context.Background(), u.String(), opt, "best")
				if dlErr != nil {
					log.Errorf("download failed: %v", dlErr)
					failures++
					continue
				}
				func() {
					defer downloadResult.Close()
					f, ferr := os.Create(downloadFilePath) //nolint:gosec // download target path chosen by user
					if ferr != nil {
						log.Errorf("create file failed: %v", ferr)
						failures++
						return
					}
					defer f.Close()
					if _, cerr := io.Copy(f, downloadResult); cerr != nil {
						log.Errorf("write file failed: %v", cerr)
						failures++
						return
					}
					_ = f.Close()
				}()

				// Remux and embed metadata (pipe policy: always)
				remuxOpt := dl.RemuxOptionsFromInfo(conf.FFmpegBin(), fs.VideoMp4, result.Info, u.String())
				if remuxErr := ffmpeg.RemuxFile(downloadFilePath, "", remuxOpt); remuxErr != nil {
					log.Errorf("remux failed: %v", remuxErr)
					failures++
					continue
				}
			} else {
				// file method
				// Deterministic output template within the session temp dir
				outTpl := filepath.Join(downloadPath, "ppdl_%(id)s.%(ext)s")
				files, dlErr := result.DownloadToFileWithOptions(context.Background(), dl.DownloadOptions{
					Filter:            "best",
					DownloadAudioOnly: false,
					EmbedMetadata:     true,
					EmbedSubs:         false,
					ForceOverwrites:   false,
					DisableCaching:    false,
					PlaylistIndex:     1,
					Output:            outTpl,
				})

				if dlErr != nil {
					log.Errorf("download failed: %v", dlErr)
					// even on error, any completed files returned will be imported
				}

				// Ensure container/metadata per remux policy for file method
				if fileRemux != "skip" {
					for _, fp := range files {
						if fileRemux == "auto" && strings.EqualFold(filepath.Ext(fp), fs.ExtMp4) {
							// Assume yt-dlp produced a valid MP4 and embedded metadata
							continue
						}
						remuxOpt := dl.RemuxOptionsFromInfo(conf.FFmpegBin(), fs.VideoMp4, result.Info, u.String())
						if remuxErr := ffmpeg.RemuxFile(fp, "", remuxOpt); remuxErr != nil {
							log.Errorf("remux failed: %v", remuxErr)
							failures++
							continue
						}
					}
				}
			}
		}
	}

	// Import results once
	log.Infof("importing downloads to %s", clean.Log(filepath.Join(conf.OriginalsPath(), destFolder)))
	w := get.Import()
	opt := photoprism.ImportOptionsMove(downloadPath, destFolder)
	w.Start(opt)

	elapsed := time.Since(start)

	if failures > 0 {
		log.Warnf("completed with %d error(s) in %s", failures, elapsed)
	} else {
		log.Infof("completed in %s", elapsed)
	}

	if failures > 0 {
		return fmt.Errorf("some downloads failed: %d", failures)
	}

	return nil
}
