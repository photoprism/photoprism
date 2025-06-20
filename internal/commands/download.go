package commands

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/dl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// DownloadCommand configures the command name, flags, and action.
var DownloadCommand = &cli.Command{
	Name:      "download",
	Aliases:   []string{"dl"},
	Usage:     "Imports media from a URL",
	ArgsUsage: "[url]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "dest",
			Aliases: []string{"d"},
			Usage:   "relative originals `PATH` to which the files should be imported",
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

	// Get URL from first argument.
	sourceUrl, sourceErr := url.Parse(strings.TrimSpace(ctx.Args().First()))

	if sourceErr != nil {
		return sourceErr
	} else if sourceUrl.Scheme != scheme.Http && sourceUrl.Scheme != scheme.Https {
		return fmt.Errorf("invalid download URL scheme %s", clean.Log(sourceUrl.Scheme))
	}

	var destFolder string
	if ctx.IsSet("dest") {
		destFolder = clean.UserPath(ctx.String("dest"))
	} else {
		destFolder = conf.ImportDest()
	}

	var downloadPath, downloadFile string

	downloadPath = filepath.Join(conf.TempPath(), "download_"+rnd.Base36(12))

	if err := fs.MkdirAll(downloadPath); err != nil {
		return err
	}

	defer os.RemoveAll(downloadPath)

	mediaType := media.FromName(sourceUrl.Path)
	mediaExt := fs.Ext(sourceUrl.Path)

	switch mediaType {
	case media.Image, media.Vector, media.Raw, media.Document, media.Audio:
		log.Infof("downloading %s from %s", mediaType, clean.Log(sourceUrl.String()))

		if dlName := clean.DlName(fs.BasePrefix(sourceUrl.Path, true)); dlName != "" {
			downloadFile = dlName + mediaExt
		} else {
			downloadFile = time.Now().Format("20060102_150405") + mediaExt
		}

		downloadFilePath := filepath.Join(downloadPath, downloadFile)

		if downloadErr := fs.Download(downloadFilePath, sourceUrl.String()); downloadErr != nil {
			return downloadErr
		}
	default:
		mediaType = media.Video
		log.Infof("downloading %s from %s", mediaType, clean.Log(sourceUrl.String()))

		opt := dl.Options{
			// The following flags currently seem to have no effect when piping the output to stdout;
			// however, that may change in a future version of the "yt-dlp" video downloader:
			MergeOutputFormat: fs.VideoMp4.String(),
			RemuxVideo:        fs.VideoMp4.String(),
			// Alternative codec sorting format to prioritize H264/AVC:
			// vcodec:h264>av01>h265>vp9.2>vp9>h263,acodec:m4a>mp4a>aac>mp3>mp3>ac3>dts
			SortingFormat: "lang,quality,res,fps,codec:avc:m4a,channels,size,br,asr,proto,ext,hasaud,source,id",
		}

		result, err := dl.NewMetadata(context.Background(), sourceUrl.String(), opt)

		if err != nil {
			return err
		}

		if dlName := clean.DlName(result.Info.Title); dlName != "" {
			downloadFile = dlName + fs.ExtMp4
		} else {
			downloadFile = time.Now().Format("20060102_150405") + fs.ExtMp4
		}

		// Compose download file path.
		downloadFilePath := filepath.Join(downloadPath, downloadFile)

		// Download the first video and embed its metadata,
		// see https://github.com/yt-dlp/yt-dlp?tab=readme-ov-file#format-selection-examples.
		downloadResult, err := result.DownloadWithOptions(context.Background(), dl.DownloadOptions{
			// TODO: While this may work with a future version of the "yt-dlp" video downloader,
			//    it is currently not possible to properly download videos with separate video and
			//    audio streams when piping the output to stdout. For now, the following Filter
			//    will download the best combined video and audio content (see docs for details).
			Filter: "best",
			// Alternative filters for combining the best video and audio streams:
			// Filter: "bestvideo*+bestaudio/best",
			// Filter: "best/bestvideo+bestaudio",
			DownloadAudioOnly: false,
			EmbedMetadata:     true,
			EmbedSubs:         false,
			ForceOverwrites:   false,
			DisableCaching:    false,
			// Download the first video if multiple videos are available:
			PlaylistIndex: 1,
		})

		// Check if download was successful.
		if err != nil {
			return err
		}

		defer downloadResult.Close()

		file, err := os.Create(downloadFilePath)

		if err != nil {
			return err
		}

		if _, err = io.Copy(file, downloadResult); err != nil {
			file.Close()
			return err
		}

		file.Close()

		// TODO: The remux command flags currently don't seem to have an effect when piping the output to stdout,
		//    so this command will manually remux the downloaded file with ffmpeg. This ensures that the file is a
		//    valid MP4 that can be played. It also adds metadata in the same step.
		remuxOpt := encode.NewRemuxOptions(conf.FFmpegBin(), fs.VideoMp4, false)

		if title := clean.Name(result.Info.Title); title != "" {
			remuxOpt.Title = title
		} else if title = clean.Name(result.Info.AltTitle); title != "" {
			remuxOpt.Title = title
		}

		if desc := strings.TrimSpace(result.Info.Description); desc != "" {
			remuxOpt.Description = desc
		}

		if u := strings.TrimSpace(sourceUrl.String()); u != "" {
			remuxOpt.Comment = u
		}

		if author := clean.Name(result.Info.Artist); author != "" {
			remuxOpt.Author = author
		} else if author = clean.Name(result.Info.AlbumArtist); author != "" {
			remuxOpt.Author = author
		} else if author = clean.Name(result.Info.Creator); author != "" {
			remuxOpt.Author = author
		} else if author = clean.Name(result.Info.License); author != "" {
			remuxOpt.Author = author
		}

		if result.Info.Timestamp > 1 {
			sec, dec := math.Modf(result.Info.Timestamp)
			remuxOpt.Created = time.Unix(int64(sec), int64(dec*(1e9)))
		}

		if remuxErr := ffmpeg.RemuxFile(downloadFilePath, "", remuxOpt); remuxErr != nil {
			return remuxErr
		}
	}

	log.Infof("importing %s to %s", mediaType, clean.Log(filepath.Join(conf.OriginalsPath(), destFolder)))

	w := get.Import()
	opt := photoprism.ImportOptionsMove(downloadPath, destFolder)

	w.Start(opt)

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
