package api

import (
	"image"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/frame"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// SharePreview returns a preview image for the given share uid if the token is valid.
//
// GET /s/:token/:shared/preview
// TODO: Proof of concept, needs refactoring.
func SharePreview(router *gin.RouterGroup) {
	router.GET("/:token/:shared/preview", func(c *gin.Context) {
		conf := get.Config()

		token := clean.Token(c.Param("token"))
		shared := clean.UID(c.Param("shared"))
		links := entity.FindLinks(token, shared)

		if len(links) != 1 {
			log.Warn("share: invalid token (preview)")
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		thumbPath := path.Join(conf.ThumbCachePath(), "share")

		if err := fs.MkdirAll(thumbPath); err != nil {
			log.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		previewFilename := filepath.Join(thumbPath, shared+fs.ExtJPEG)

		expires := entity.Now().Add(-1 * time.Hour)

		if info, err := os.Stat(previewFilename); err != nil {
			log.Debugf("share: creating new preview for %s", clean.Log(shared))
		} else if info.ModTime().After(expires) {
			log.Debugf("share: using cached preview for %s", clean.Log(shared))
			c.File(previewFilename)
			return
		} else if err := os.Remove(previewFilename); err != nil {
			log.Errorf("share: could not remove old preview of %s", clean.Log(shared))
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		a, err := query.AlbumByUID(shared)

		if err != nil {
			log.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		var f form.SearchPhotos

		// Covers may only contain public content in shared albums.
		f.Album = shared
		f.Public = true
		f.Private = false
		f.Hidden = false
		f.Archived = false
		f.Review = false
		f.Primary = true

		// Get first 12 album entries.
		f.Count = 6
		f.Order = a.AlbumOrder

		if parseErr := f.ParseQueryString(); parseErr != nil {
			log.Errorf("preview: %s", parseErr)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		p, count, err := search.Photos(f)

		if err != nil {
			log.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		if count == 0 {
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		size, _ := thumb.Sizes[thumb.Tile500]

		images := make([]image.Image, 0, len(p))

		// Get thumbnail images to create album preview.
		for _, file := range p {
			fileName := photoprism.FileName(file.FileRoot, file.FileName)

			if !fs.FileExists(fileName) {
				log.Errorf("share: file %s is missing (preview)", clean.Log(file.FileName))
				c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
				return
			}

			thumbnail, imgErr := thumb.FromFile(fileName, file.FileHash, conf.ThumbCachePath(), size.Width, size.Height, file.FileOrientation, size.Options...)

			if imgErr != nil {
				log.Warn(imgErr)
				continue
			}

			img, imgErr := imaging.Open(thumbnail)

			if imgErr != nil {
				log.Warn(imgErr)
				continue
			}

			images = append(images, img)
		}

		// Create album preview from thumbnail images.
		preview, err := frame.Collage(frame.Polaroid, images)

		// Downsize from 1600x900 to 1200x675.
		preview = imaging.Resize(preview, 1200, 0, imaging.Lanczos)

		// Save the resulting album preview as JPEG.
		err = imaging.Save(preview, previewFilename, thumb.JpegQualitySmall().EncodeOption())

		if err != nil {
			log.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		c.File(previewFilename)
	})
}
