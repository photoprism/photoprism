package api

import (
	"image"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/frame"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// SharePreview returns a preview image for the given share uid if the token is valid.
//
//	@Summary	returns a share preview image when the token is valid
//	@Id			SharePreview
//	@Tags		Sharing
//	@Produce	image/jpeg
//	@Param		token	path		string	true	"Share token"
//	@Param		shared	path		string	true	"Shared resource UID"
//	@Success	200		{file}		file	"Preview image"
//	@Failure	302		{string}	string	"Redirect to the default preview page"
//	@Router		/s/{token}/{shared}/preview [get]
func SharePreview(router *gin.RouterGroup) {
	router.GET("/:token/:shared/preview", func(c *gin.Context) {
		conf := get.Config()

		token := clean.ShareToken(c.Param("token"))
		shared := clean.UID(c.Param("shared"))
		links := entity.FindRedeemableLinksByToken(token, shared)

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

		previewFilename := filepath.Join(thumbPath, shared+fs.ExtJpeg)

		expires := entity.Now().Add(-1 * time.Hour)

		if info, err := os.Stat(previewFilename); err != nil {
			log.Debugf("share: creating new preview for %s", clean.Log(shared))
		} else if info.ModTime().After(expires) {
			// An empty file marks an album that composed no image, so that answer is as cheap to
			// repeat as a cached card and expires on the same schedule.
			if info.Size() == 0 {
				c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
				return
			}

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

		var frm form.SearchPhotos

		// Covers may only contain public content in shared albums. SharedPhotos below applies
		// the same five constraints after any smart-album filter, so the values here are the
		// request's starting point rather than the boundary.
		frm.Album = shared
		frm.Public = true
		frm.Private = false
		frm.Hidden = false
		frm.Archived = false
		frm.Review = false
		frm.Primary = true

		// Get first 12 album entries.
		frm.Count = 6
		frm.Order = a.AlbumOrder

		if parseErr := frm.ParseQueryString(); parseErr != nil {
			log.Errorf("preview: %s", parseErr)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		p, count, err := search.SharedPhotos(frm)

		if err != nil {
			log.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		if count == 0 {
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		size := thumb.Sizes[thumb.Tile500]

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

			img, _, imgErr := fs.DecodeImageFile(thumbnail)

			if imgErr != nil {
				log.Warn(imgErr)
				continue
			}

			images = append(images, img)
		}

		// A selection that yields no image has nothing to compose, so the request serves the site
		// preview and marks the album with an empty file for the lifetime of a preview.
		if len(images) == 0 {
			log.Debugf("share: no image to compose for %s", clean.Log(shared))
			markEmptyPreview(previewFilename)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		// Create album preview from thumbnail images.
		preview, err := frame.Collage(frame.Polaroid, images)
		if err != nil {
			log.Warnf("preview collage: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		// Downsize from 1600x900 to 1200x675.
		preview = thumb.Resample(preview, 1200, 675, thumb.ResampleResize)

		// Save the resulting album preview as JPEG.
		err = thumb.Save(preview, previewFilename, thumb.JpegQualitySmall())

		if err != nil {
			log.Error(err)
			c.Redirect(http.StatusTemporaryRedirect, conf.SitePreview())
			return
		}

		c.File(previewFilename)
	})
}

// markEmptyPreview creates the empty file that marks an album as composing no preview image. The
// name is claimed exclusively, so only a call that finds it free writes the marker.
func markEmptyPreview(fileName string) {
	// #nosec G304 -- the name is a validated UID under the thumbnail cache.
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fs.ModeFile)

	if err != nil {
		log.Debugf("share: %s (mark preview)", clean.Error(err))
		return
	}

	if err = f.Close(); err != nil {
		log.Debugf("share: %s (mark preview)", clean.Error(err))
	}
}
