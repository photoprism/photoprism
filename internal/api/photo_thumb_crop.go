package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/crop"
	"github.com/photoprism/photoprism/internal/service"
)

// GetThumbCrop returns a cropped thumbnail image matching the hash and type.
//
// GET /api/v1/t/:hash/:token/:size/:crop
//
// Parameters:
//   hash: string sha1 file hash
//   token: string url security token, see config
//   size: string crop size, see crop.Sizes
//   area: string image area identifier, e.g. 1690960ff17f
func GetThumbCrop(router *gin.RouterGroup) {
	router.GET("/t/:hash/:token/:size/:area", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		logPrefix := "thumb-crop"

		conf := service.Config()
		fileHash := c.Param("hash")
		cropName := crop.Name(c.Param("size"))
		cropArea := c.Param("area")
		download := c.Query("download") != ""

		cropSize, ok := crop.Sizes[cropName]

		if !ok {
			log.Errorf("%s: invalid size %s", logPrefix, cropName)
			c.Data(http.StatusOK, "image/svg+xml", photoIconSvg)
			return
		}

		fileName, err := crop.FromCache(fileHash, cropArea, cropSize, conf.ThumbPath())

		if err != nil {
			log.Warnf("%s: %s", logPrefix, err)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		} else if fileName == "" {
			log.Errorf("%s: empty file name, potential bug", logPrefix)
			c.Data(http.StatusOK, "image/svg+xml", brokenIconSvg)
			return
		}

		AddThumbCacheHeader(c)

		if download {
			c.FileAttachment(fileName, cropName.Jpeg())
		} else {
			c.File(fileName)
		}
	})
}
