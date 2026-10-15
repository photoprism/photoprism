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

// runDownload executes the download/import flow for the given inputs and options; downloadAction calls it.
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

	// Downloads are imported, so the destination must give names that the index accepts.
	if err := get.Import().CheckDest(destFolder); err != nil {
		return err
	}

	// Create session download directory
	downloadPath := filepath.Join(conf.TempPath(), fs.DownloadDir+"_"+rnd.Base36(12))
	if err := fs.MkdirAll(downloadPath); err != nil {
		return err
	}
	defer func() {
		if rmErr := os.RemoveAll(downloadPath); rmErr != nil {
			log.Debugf("download: %s (remove temporary download path)", clean.ErrorFull(rmErr))
		}
	}()

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
		return fmt.Errorf("invalid --remux: %s (expected 'always', 'auto', or 'skip')", fileRemux)
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
			log.Errorf("invalid URL: %s", clean.LogUri(raw))
			failures++
			continue
		}
		if u.Scheme != scheme.Http && u.Scheme != scheme.Https {
			log.Errorf("invalid URL scheme %s: %s", clean.Log(u.Scheme), clean.LogUri(raw))
			failures++
			continue
		}

		mt := media.FromName(u.Path)
		ext := fs.Ext(u.Path)
		var downloadFile string

		switch mt {
		case media.Image, media.Vector, media.Raw, media.Document, media.Audio:
			log.Infof("downloading %s from %s", mt, clean.LogUri(u.String()))
			downloadFile = downloadName(downloadPath, fs.BasePrefix(u.Path, true), ext)
			downloadFilePath := filepath.Join(downloadPath, downloadFile)
			if downloadErr := fs.Download(downloadFilePath, u.String()); downloadErr != nil {
				log.Errorf("download failed: %s", clean.ErrorFull(downloadErr))
				failures++
				continue
			}
		default:
			mt = media.Video
			log.Infof("downloading %s from %s", mt, clean.LogUri(u.String()))
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
				log.Errorf("metadata failed: %s", clean.ErrorFull(err))
				if hint, ok := missingFormatsHint(err); ok {
					log.Info(hint)
				}
				failures++
				continue
			}

			if matched := ffmpeg.Exclude().Match(result.Info.VCodec, result.Info.Ext, result.Info.Container); matched != "" {
				log.Warnf("skipping %s because format %s is on the FFmpeg exclude list", clean.LogUri(u.String()), clean.Log(matched))
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
			downloadFile = downloadName(downloadPath, result.Info.Title, fs.ExtMp4)
			downloadFilePath := filepath.Join(downloadPath, downloadFile)

			if method == "pipe" {
				downloadResult, err := dl.Download(context.Background(), u.String(), opt, dl.FormatBest)
				if err != nil {
					if downloadResult != nil {
						_ = downloadResult.Close()
					}
					log.Errorf("download failed: %s", clean.ErrorFull(err))
					failures++
					continue
				}
				if !savePipeDownload(downloadResult, downloadFilePath) {
					failures++
					continue
				}

				remuxOpt := dl.RemuxOptionsFromInfo(conf.FFmpegBin(), fs.VideoMp4, result.Info, u.String())
				if remuxErr := ffmpeg.RemuxFile(downloadFilePath, "", remuxOpt); remuxErr != nil {
					log.Errorf("remux failed: %s", clean.ErrorFull(remuxErr))
					failures++
					continue
				}
			} else {
				outTpl := filepath.Join(downloadPath, "ppdl_%(id)s.%(ext)s")
				files, err := result.DownloadToFileWithOptions(context.Background(), dl.DownloadOptions{
					Filter:            dl.FormatBest,
					DownloadAudioOnly: false,
					EmbedMetadata:     true,
					EmbedSubs:         false,
					ForceOverwrites:   false,
					DisableCaching:    false,
					PlaylistIndex:     1,
					Output:            outTpl,
				})
				if err != nil {
					log.Errorf("download failed: %s", clean.ErrorFull(err))
					if len(files) == 0 {
						failures++
						continue
					}
				}
				if fileRemux != "skip" {
					for _, fp := range files {
						if fileRemux == "auto" && strings.EqualFold(filepath.Ext(fp), fs.ExtMp4) {
							continue
						}
						remuxOpt := dl.RemuxOptionsFromInfo(conf.FFmpegBin(), fs.VideoMp4, result.Info, u.String())
						if remuxErr := ffmpeg.RemuxFile(fp, "", remuxOpt); remuxErr != nil {
							log.Errorf("remux failed: %s", clean.ErrorFull(remuxErr))
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
		log.Warnf("completed with %s in %s", formatCount(failures, "error", "errors"), elapsed)
		return fmt.Errorf("%s", formatFailedCount(failures, "download", "downloads"))
	}
	log.Infof("completed in %s", elapsed)
	return nil
}

// savePipeDownload writes a pipe download to the file path and removes the file if the stream
// failed, so an incomplete download is neither remuxed nor imported.
func savePipeDownload(stream io.ReadCloser, filePath string) bool {
	f, err := os.Create(filePath) //nolint:gosec // download target path chosen by user
	if err != nil {
		if closeErr := stream.Close(); closeErr != nil {
			log.Debugf("download: %s (close stream after create failure)", clean.ErrorFull(closeErr))
		}
		log.Errorf("create file failed: %s", clean.ErrorFull(err))
		return false
	}

	_, copyErr := io.Copy(f, stream)
	closeFileErr := f.Close()
	closeStreamErr := stream.Close()

	switch {
	case copyErr != nil:
		log.Errorf("write file failed: %s", clean.ErrorFull(copyErr))
	case closeFileErr != nil:
		log.Errorf("close file failed: %s", clean.ErrorFull(closeFileErr))
	case closeStreamErr != nil:
		log.Errorf("download failed: %s", clean.ErrorFull(closeStreamErr))
	default:
		return true
	}

	if rmErr := os.Remove(filePath); rmErr != nil {
		log.Debugf("download: %s (remove incomplete file)", clean.ErrorFull(rmErr))
	}

	return false
}

// missingFormatsHint returns a hint for yt-dlp errors that indicate missing formats.
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

// downloadName returns a file name in dir that is not used yet for a download from a title or remote name, or a
// time-based name if nothing usable is left of it, or the result is not portable and would therefore not be imported.
func downloadName(dir, name, ext string) string {
	base := clean.DlName(name)

	if base == "" || !clean.PortableName(base+ext) {
		base = time.Now().Format("20060102_150405")
	}

	result := base + ext

	for i := 2; fs.FileExists(filepath.Join(dir, result)) || fs.IsSymlink(filepath.Join(dir, result)); i++ {
		result = fmt.Sprintf("%s_%d%s", base, i, ext)
	}

	return result
}
