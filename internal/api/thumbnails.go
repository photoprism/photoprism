package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism"
	"log"
	"strconv"
)

var photoIconSvg = []byte(`
<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24">
<path d="M0 0h24v24H0z" fill="none"/>
<path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
</svg>`)

// GetThumbnail searches the database for a thumbnail based on hash, size, and type.
func GetThumbnail(router *gin.RouterGroup, conf *photoprism.Config) {
	router.GET("/thumbnails/:type/:size/:hash", func(c *gin.Context) {
		fileHash := c.Param("hash")
		thumbnailType := c.Param("type")
		size, err := strconv.Atoi(c.Param("size"))

		if err != nil {
			log.Printf("invalid size: %s", c.Param("size"))
			c.Data(400, "image/svg+xml", photoIconSvg)
		}

		search := photoprism.NewSearch(conf.OriginalsPath, conf.GetDb())

		file := search.FindFileByHash(fileHash)

		fileName := fmt.Sprintf("%s/%s", conf.OriginalsPath, file.FileName)

		if mediaFile, err := photoprism.NewMediaFile(fileName); err == nil {
			switch thumbnailType {
			case "fit":
				if thumbnail, err := mediaFile.GetThumbnail(conf.ThumbnailsPath, size); err == nil {
					c.File(thumbnail.GetFilename())
				} else {
					log.Printf("could not create thumbnail: %s", err.Error())
					c.Data(400, "image/svg+xml", photoIconSvg)
				}
			case "square":
				if thumbnail, err := mediaFile.GetSquareThumbnail(conf.ThumbnailsPath, size); err == nil {
					c.File(thumbnail.GetFilename())
				} else {
					log.Printf("could not create square thumbnail: %s", err.Error())
					c.Data(400, "image/svg+xml", photoIconSvg)
				}
			default:
				log.Printf("unknown thumbnail type: %s", thumbnailType)
				c.Data(400, "image/svg+xml", photoIconSvg)
			}
		} else {
			log.Printf("could not find image for thumbnail: %s", err.Error())
			c.Data(404, "image/svg+xml", photoIconSvg)

			// Set missing flag so that the file doesn't show up in search results anymore
			file.FileMissing = true
			conf.GetDb().Save(&file)
		}
	})
}
