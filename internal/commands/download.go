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

	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/photoprism/ytdl"
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

		result, err := ytdl.New(context.Background(), sourceUrl.String(), ytdl.Options{})
		if err != nil {
			return err
		}

		if dlName := clean.DlName(result.Info.Title); dlName != "" {
			downloadFile = dlName + fs.ExtMp4
		} else {
			downloadFile = time.Now().Format("20060102_150405") + fs.ExtMp4
		}

		downloadFilePath := filepath.Join(downloadPath, downloadFile)
		downloadResult, err := result.Download(context.Background(), "best")

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
	}

	log.Infof("importing %s to %s", mediaType, clean.Log(filepath.Join(conf.OriginalsPath(), destFolder)))

	w := get.Import()
	opt := photoprism.ImportOptionsMove(downloadPath, destFolder)

	w.Start(opt)

	elapsed := time.Since(start)

	log.Infof("completed in %s", elapsed)

	return nil
}
