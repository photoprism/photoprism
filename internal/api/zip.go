package api

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// CreateZip creates a zip file archive for download.
//
// POST /api/v1/zip
func CreateZip(router *gin.RouterGroup) {
	router.POST("/zip", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionDownload)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		conf := service.Config()

		if !conf.Settings().Features.Download {
			AbortFeatureDisabled(c)
			return
		}

		var f form.Selection
		start := time.Now()

		if err := c.BindJSON(&f); err != nil {
			AbortBadRequest(c)
			return
		}

		if f.Empty() {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		// Configure file selection based on user settings.
		var selection query.FileSelection
		if dl := conf.Settings().Download; dl.Disabled {
			AbortFeatureDisabled(c)
			return
		} else {
			selection = query.DownloadSelection(dl.MediaRaw, dl.MediaSidecar, dl.Originals)
		}

		// Find files to download.
		files, err := query.SelectedFiles(f, selection)

		if err != nil {
			Error(c, http.StatusBadRequest, err, i18n.ErrZipFailed)
			return
		} else if len(files) == 0 {
			Abort(c, http.StatusNotFound, i18n.ErrNoFilesForDownload)
			return
		}

		// Configure file names.
		dlName := DownloadName(c)
		zipPath := path.Join(conf.TempPath(), "zip")
		zipToken := rnd.GenerateToken(8)
		zipBaseName := fmt.Sprintf("photoprism-download-%s-%s.zip", time.Now().Format("20060102-150405"), zipToken)
		zipFileName := path.Join(zipPath, zipBaseName)

		// Create temp directory.
		if err = os.MkdirAll(zipPath, 0700); err != nil {
			Error(c, http.StatusInternalServerError, err, i18n.ErrZipFailed)
			return
		}

		// Create new zip file.
		var newZipFile *os.File
		if newZipFile, err = os.Create(zipFileName); err != nil {
			Error(c, http.StatusInternalServerError, err, i18n.ErrZipFailed)
			return
		} else {
			defer newZipFile.Close()
		}

		// Create zip writer.
		zipWriter := zip.NewWriter(newZipFile)
		defer func(w *zip.Writer) {
			logError("zip", w.Close())
		}(zipWriter)

		var aliases = make(map[string]int)

		// Add files to zip.
		for _, file := range files {
			fileName := photoprism.FileName(file.FileRoot, file.FileName)
			alias := file.DownloadName(dlName, 0)
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.DownloadName(dlName, seq)
			}

			aliases[key] += 1

			if fs.FileExists(fileName) {
				if err := addFileToZip(zipWriter, fileName, alias); err != nil {
					log.Errorf("download: failed adding %s to zip (%s)", clean.Log(file.FileName), err)
					Abort(c, http.StatusInternalServerError, i18n.ErrZipFailed)
					return
				}

				log.Infof("download: added %s as %s", clean.Log(file.FileName), clean.Log(alias))
			} else {
				log.Warnf("download: media file %s is missing", clean.Log(file.FileName))
				logError("download", file.Update("FileMissing", true))
			}
		}

		elapsed := int(time.Since(start).Seconds())

		log.Infof("download: created %s [%s]", clean.Log(zipBaseName), time.Since(start))

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgZipCreatedIn, elapsed), "filename": zipBaseName})
	})
}

// DownloadZip downloads a zip file archive.
//
// GET /api/v1/zip/:filename
func DownloadZip(router *gin.RouterGroup) {
	router.GET("/zip/:filename", func(c *gin.Context) {
		if InvalidDownloadToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		conf := service.Config()
		zipBaseName := clean.FileName(filepath.Base(c.Param("filename")))
		zipPath := path.Join(conf.TempPath(), "zip")
		zipFileName := path.Join(zipPath, zipBaseName)

		if !fs.FileExists(zipFileName) {
			log.Errorf("could not find zip file: %s", clean.Log(zipFileName))
			c.Data(404, "image/svg+xml", photoIconSvg)
			return
		}

		c.FileAttachment(zipFileName, zipBaseName)

		defer func(n string) {
			logError("zip", os.Remove(n))
		}(zipFileName)
	})
}

// addFileToZip adds a file to a zip archive.
func addFileToZip(zipWriter *zip.Writer, fileName, fileAlias string) error {
	fileToZip, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = fileAlias

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
