package api

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
)

// POST /api/v1/zip
func CreateZip(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/zip", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		if !conf.Settings().Features.Download {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrFeatureDisabled)
			return
		}

		var f form.Selection
		start := time.Now()

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if f.Empty() {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst("no items selected")})
			return
		}

		files, err := query.FileSelection(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		} else if len(files) == 0 {
			c.AbortWithStatusJSON(404, gin.H{"error": txt.UcFirst("no files available for download")})
			return
		}

		zipPath := path.Join(conf.TempPath(), "zip")
		zipToken := rnd.Token(3)
		zipYear := time.Now().Format("January-2006")
		zipBaseName := fmt.Sprintf("Photos-%s-%s.zip", zipYear, zipToken)
		zipFileName := path.Join(zipPath, zipBaseName)

		if err := os.MkdirAll(zipPath, 0700); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst("failed to create zip folder")})
			return
		}

		newZipFile, err := os.Create(zipFileName)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		defer newZipFile.Close()

		zipWriter := zip.NewWriter(newZipFile)
		defer zipWriter.Close()

		for _, f := range files {
			fileName := path.Join(conf.OriginalsPath(), f.FileName)
			fileAlias := f.ShareFileName()

			if fs.FileExists(fileName) {
				if err := addFileToZip(zipWriter, fileName, fileAlias); err != nil {
					log.Error(err)
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": txt.UcFirst("failed to create zip file")})
					return
				}
				log.Infof("zip: added %s as %s", txt.Quote(f.FileName), txt.Quote(fileAlias))
			} else {
				log.Warnf("zip: file %s is missing", txt.Quote(f.FileName))
				logError("zip", f.Update("FileMissing", true))
			}
		}

		elapsed := int(time.Since(start).Seconds())

		log.Infof("zip: archive %s created in %s", txt.Quote(zipBaseName), time.Since(start))

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("zip created in %d s", elapsed), "filename": zipBaseName})
	})
}

// GET /api/v1/zip/:filename
func DownloadZip(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/zip/:filename", func(c *gin.Context) {
		if InvalidDownloadToken(c, conf) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		zipBaseName := filepath.Base(c.Param("filename"))
		zipPath := path.Join(conf.TempPath(), "zip")
		zipFileName := path.Join(zipPath, zipBaseName)

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipBaseName))

		if !fs.FileExists(zipFileName) {
			log.Errorf("could not find zip file: %s", zipFileName)
			c.Data(404, "image/svg+xml", photoIconSvg)
			return
		}

		c.File(zipFileName)

		if err := os.Remove(zipFileName); err != nil {
			log.Errorf("zip: could not remove %s (%s)", txt.Quote(zipFileName), err.Error())
		}
	})
}

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
