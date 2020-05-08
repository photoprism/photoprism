package api

import (
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

// GET /api/v1/preview
func GetPreview(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/preview", func(c *gin.Context) {
		// TODO: proof of concept - code needs refactoring!
		t := time.Now().Format("20060102")
		thumbPath := path.Join(conf.ThumbPath(), "preview", t[0:4], t[4:6])

		if err := os.MkdirAll(thumbPath, os.ModePerm); err != nil {
			log.Error(err)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		previewFilename := fmt.Sprintf("%s/%s.jpg", thumbPath, t[6:8])

		if fs.FileExists(previewFilename) {
			c.File(previewFilename)
			return
		}

		var f form.PhotoSearch

		f.Public = true
		f.Safe = true
		f.Count = 12
		f.Order = "relevance"

		p, _, err := query.Photos(f)

		if err != nil {
			log.Error(err)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		width := 908
		height := 680
		x := 0
		y := 0

		preview := imaging.New(width, height, color.NRGBA{255, 255, 255, 255})
		thumbType, _ := thumb.Types["tile_224"]

		for _, f := range p {
			fileName := path.Join(conf.OriginalsPath(), f.FileName)

			if !fs.FileExists(fileName) {
				log.Errorf("could not find original for thumbnail: %s", fileName)
				c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)

				// Set missing flag so that the file doesn't show up in search results anymore
				f.FileMissing = true
				conf.Db().Save(&f)
				return
			}

			thumbnail, err := thumb.FromFile(fileName, f.FileHash, conf.ThumbPath(), thumbType.Width, thumbType.Height, thumbType.Options...)

			if err != nil {
				log.Error(err)
				c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			}

			src, err := imaging.Open(thumbnail)

			if err != nil {
				log.Error(err)
				c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			}

			preview = imaging.Paste(preview, src, image.Pt(x, y))

			x += 228

			if x > width {
				x = 0
				y += 228
			}
		}

		// Save the resulting image as JPEG.
		err = imaging.Save(preview, previewFilename)

		if err != nil {
			log.Error(err)
			c.Data(http.StatusNotFound, "image/svg+xml", photoIconSvg)
			return
		}

		c.File(previewFilename)
	})
}
