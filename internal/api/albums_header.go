package api

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// AlbumHeaderData describes a published cover without exposing photo or person metadata.
type AlbumHeaderData struct {
	Hash   string           `json:"Hash"`
	Width  int              `json:"Width"`
	Height int              `json:"Height"`
	Faces  *AlbumHeaderRect `json:"Faces,omitempty"`
}

// AlbumHeaderRect encloses all valid face regions in normalized image coordinates.
type AlbumHeaderRect struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	W float64 `json:"W"`
	H float64 `json:"H"`
}

// GetAlbumHeader returns the published cover and anonymous framing information.
//
//	@Summary	returns album header information
//	@Id			GetAlbumHeader
//	@Tags		Albums
//	@Produce	json
//	@Success	200				{object}	AlbumHeaderData
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		uid				path		string	true	"Album UID"
//	@Router		/api/v1/albums/{uid}/header [get]
func GetAlbumHeader(router *gin.RouterGroup) {
	router.GET("/albums/:uid/header", func(c *gin.Context) {
		s := Auth(c, acl.ResourceAlbums, acl.ActionView)
		if s.Abort(c) {
			return
		}
		album, err := query.AlbumByUID(clean.UID(c.Param("uid")))
		if err != nil || !albumViewableBySession(s, album) {
			AbortAlbumNotFound(c)
			return
		}

		c.Header("Cache-Control", "no-store")
		var file *entity.File
		if album.Thumb != "" {
			file, err = query.FileByHash(album.Thumb)
		} else {
			var resolved entity.File
			resolved, err = query.AlbumCoverByUID(album.AlbumUID, get.Config().Settings().Features.Private)
			file = &resolved
		}
		if err != nil || file == nil || file.FileHash == "" {
			c.JSON(http.StatusOK, AlbumHeaderData{})
			return
		}

		var markers entity.Markers
		// Only geometry is read; identities and embeddings never enter the response.
		if err = entity.Db().Select("x, y, w, h").Where("file_uid = ? AND marker_type = ? AND marker_invalid = FALSE", file.FileUID, entity.MarkerFace).Find(&markers).Error; err != nil {
			AbortUnexpectedError(c)
			return
		}
		c.JSON(http.StatusOK, AlbumHeaderData{Hash: file.FileHash, Width: file.FileWidth, Height: file.FileHeight, Faces: albumHeaderFaces(markers)})
	})
}

// albumHeaderFaces returns the clamped union of valid face rectangles.
func albumHeaderFaces(markers entity.Markers) *AlbumHeaderRect {
	left, top, right, bottom := 1.0, 1.0, 0.0, 0.0
	for _, m := range markers {
		x, y, w, h := float64(m.X), float64(m.Y), float64(m.W), float64(m.H)
		if w <= 0 || h <= 0 || math.IsNaN(x+y+w+h) || math.IsInf(x+y+w+h, 0) {
			continue
		}
		left, top = math.Min(left, math.Max(0, x)), math.Min(top, math.Max(0, y))
		right, bottom = math.Max(right, math.Min(1, x+w)), math.Max(bottom, math.Min(1, y+h))
	}
	if right <= left || bottom <= top {
		return nil
	}
	return &AlbumHeaderRect{X: left, Y: top, W: right - left, H: bottom - top}
}
