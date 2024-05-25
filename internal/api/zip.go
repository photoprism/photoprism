package api

import (
	"archive/zip"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ZipCreate creates a zip file archive for download.
//
// POST /api/v1/zip
func ZipCreate(router *gin.RouterGroup) {
	router.POST("/zip", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionDownload)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.Settings().Features.Download {
			AbortFeatureDisabled(c)
			return
		}

		var f form.Selection
		start := time.Now()

		// Assign and validate request form values.
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
		zipToken := rnd.Base36(8)
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
			logErr("zip", w.Close())
		}(zipWriter)

		var aliases = make(map[string]int)

		// Add files to zip.
		for _, file := range files {
			if file.FileName == "" {
				log.Warnf("download: %s cannot be downloaded (empty file name)", clean.Log(file.FileUID))
				continue
			} else if file.FileHash == "" {
				log.Warnf("download: %s cannot be downloaded (empty file hash)", clean.Log(file.FileName))
				continue
			}

			fileName := photoprism.FileName(file.FileRoot, file.FileName)
			alias := file.DownloadName(dlName, 0)
			key := strings.ToLower(alias)

			if seq := aliases[key]; seq > 0 {
				alias = file.DownloadName(dlName, seq)
			}

			aliases[key] += 1

			if fs.FileExists(fileName) {
				if zipErr := fs.ZipFile(zipWriter, fileName, alias, false); zipErr != nil {
					log.Errorf("download: failed to add %s (%s)", clean.Log(file.FileName), zipErr)
					Abort(c, http.StatusInternalServerError, i18n.ErrZipFailed)
					return
				}

				log.Infof("download: added %s as %s", clean.Log(file.FileName), clean.Log(alias))
			} else {
				log.Warnf("download: %s not found", clean.Log(file.FileName))
				logErr("download", file.Update("FileMissing", true))
			}
		}

		elapsed := int(time.Since(start).Seconds())

		log.Infof("download: created %s [%s]", clean.Log(zipBaseName), time.Since(start))

		c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": i18n.Msg(i18n.MsgZipCreatedIn, elapsed), "filename": zipBaseName})
	})
}

// ZipDownload downloads a zip file archive.
//
// GET /api/v1/zip/:filename
func ZipDownload(router *gin.RouterGroup) {
	router.GET("/zip/:filename", func(c *gin.Context) {
		if InvalidDownloadToken(c) {
			log.Errorf("download: %s", c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid download token")))
			return
		}

		conf := get.Config()
		zipBaseName := clean.FileName(filepath.Base(c.Param("filename")))
		zipPath := path.Join(conf.TempPath(), "zip")
		zipFileName := path.Join(zipPath, zipBaseName)

		if !fs.FileExists(zipFileName) {
			log.Errorf("download: %s", c.AbortWithError(http.StatusNotFound, fmt.Errorf("%s not found", clean.Log(zipFileName))))
			return
		}

		defer func(fileName, baseName string) {
			log.Infof("download: %s has been downloaded", clean.Log(baseName))

			// Wait a moment before deleting the zip file, just to be sure:
			// https://github.com/photoprism/photoprism/issues/2532
			time.Sleep(time.Second)

			// Remove the zip file to free up disk space.
			if err := os.Remove(fileName); err != nil {
				log.Warnf("download: failed to delete %s (%s)", clean.Log(fileName), err)
			} else {
				log.Debugf("download: deleted %s", clean.Log(baseName))
			}
		}(zipFileName, zipBaseName)

		c.FileAttachment(zipFileName, zipBaseName)
	})
}
