package commands

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

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

const pipeSortingFormat = "lang,quality,res,fps,codec:avc:m4a,channels,size,br,asr,proto,ext,hasaud,source,id"

// DownloadOpts contains the command options used by runDownload.
type DownloadOpts struct {
	Dest               string
	Cookies            string
	CookiesFromBrowser string
	AddHeaders         []string
	Impersonate        string
	Method             string // pipe|file
	FileRemux          string // always|auto|skip
	FormatSort         string
}

// runDownload executes the download/import flow for the given inputs and options.
// It is the testable core used by the CLI action.
func runDownload(conf *config.Config, opts DownloadOpts, inputURLs []string) error {
	start := time.Now()
	if conf == nil {
		return fmt.Errorf("nil config")
	}
	if conf.ReadOnly() {
		return config.ErrReadOnly
	}
	if len(inputURLs) == 0 {
		return fmt.Errorf("no download URLs provided")
	}

	if msg, ok := dl.VersionWarning(); ok {
		log.Info(msg)
	}

	// Resolve destination folder
	destFolder := opts.Dest
	if destFolder == "" {
		destFolder = conf.ImportDest()
	} else {
		destFolder = clean.UserPath(destFolder)
	}

	// Create session download directory
	downloadPath := filepath.Join(conf.TempPath(), fs.DownloadDir+"_"+rnd.Base36(12))
	if err := fs.MkdirAll(downloadPath); err != nil {
		return err
	}
	defer os.RemoveAll(downloadPath)

	// Normalize method/remux policy
	method, _, err := resolveDownloadMethod(opts.Method)
	if err != nil {
		return err
	}

	sortingFormat := strings.TrimSpace(opts.FormatSort)

	if sortingFormat == "" && method == "pipe" {
		sortingFormat = pipeSortingFormat
	}

	fileRemux := strings.ToLower(strings.TrimSpace(opts.FileRemux))

	if fileRemux == "" {
		fileRemux = "auto"
	}

	switch fileRemux {
	case "always", "auto", "skip":
	default:
		return fmt.Errorf("invalid file remux policy: %s", fileRemux)
	}

	impersonate := strings.TrimSpace(opts.Impersonate)
	if impersonate == "" {
		impersonate = "firefox"
	}
	if strings.EqualFold(impersonate, "none") {
		impersonate = ""
	} else {
		impersonate = strings.ToLower(impersonate)
	}

	// Process inputs sequentially
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
		var downloadFile string

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
				SortingFormat:      sortingFormat,
				Cookies:            opts.Cookies,
				CookiesFromBrowser: opts.CookiesFromBrowser,
				AddHeaders:         opts.AddHeaders,
				Impersonate:        impersonate,
			}
			ytRemux := method != "pipe"
			if ytRemux {
				opt.MergeOutputFormat = fs.VideoMp4.String()
				opt.RemuxVideo = fs.VideoMp4.String()
			}
			result, err := dl.NewMetadata(context.Background(), u.String(), opt)
			if err != nil {
				log.Errorf("metadata failed: %v", err)
				if hint, ok := missingFormatsHint(err); ok {
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
			if dlName := clean.DlName(result.Info.Title); dlName != "" {
				downloadFile = dlName + fs.ExtMp4
			} else {
				downloadFile = time.Now().Format("20060102_150405") + fs.ExtMp4
			}
			downloadFilePath := filepath.Join(downloadPath, downloadFile)

			if method == "pipe" {
				downloadResult, err := dl.Download(context.Background(), u.String(), opt, "best")
				if err != nil {
					log.Errorf("download failed: %v", err)
					failures++
					continue
				}
				func() {
					defer downloadResult.Close()
					f, ferr := os.Create(downloadFilePath)
					if ferr != nil {
						log.Errorf("create file failed: %v", ferr)
						failures++
						return
					}
					if _, cerr := io.Copy(f, downloadResult); cerr != nil {
						_ = f.Close()
						log.Errorf("write file failed: %v", cerr)
						failures++
						return
					}
					_ = f.Close()
				}()

				remuxOpt := dl.RemuxOptionsFromInfo(conf.FFmpegBin(), fs.VideoMp4, result.Info, u.String())
				if remuxErr := ffmpeg.RemuxFile(downloadFilePath, "", remuxOpt); remuxErr != nil {
					log.Errorf("remux failed: %v", remuxErr)
					failures++
					continue
				}
			} else {
				outTpl := filepath.Join(downloadPath, "ppdl_%(id)s.%(ext)s")
				files, err := result.DownloadToFileWithOptions(context.Background(), dl.DownloadOptions{
					Filter:            "best",
					DownloadAudioOnly: false,
					EmbedMetadata:     true,
					EmbedSubs:         false,
					ForceOverwrites:   false,
					DisableCaching:    false,
					PlaylistIndex:     1,
					Output:            outTpl,
				})
				if err != nil {
					log.Errorf("download failed: %v", err)
				}
				if fileRemux != "skip" {
					for _, fp := range files {
						if fileRemux == "auto" && strings.EqualFold(filepath.Ext(fp), fs.ExtMp4) {
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

	log.Infof("importing downloads to %s", clean.Log(filepath.Join(conf.OriginalsPath(), destFolder)))
	w := get.Import()
	opt := photoprism.ImportOptionsMove(downloadPath, destFolder)
	w.Start(opt)

	elapsed := time.Since(start)
	if failures > 0 {
		log.Warnf("completed with %d error(s) in %s", failures, elapsed)
		return fmt.Errorf("some downloads failed: %d", failures)
	}
	log.Infof("completed in %s", elapsed)
	return nil
}

func missingFormatsHint(err error) (string, bool) {
	if err == nil {
		return "", false
	}

	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "requested format is not available") {
		return "yt-dlp did not receive playable formats. Try downloading via yt-dlp --list-formats, or pass authenticated cookies with --cookies <file> so YouTube exposes video/audio streams.", true
	}

	return "", false
}
