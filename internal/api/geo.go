package api

import (
	"net/http"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/form"

	geojson "github.com/paulmach/go.geojson"
)

// GET /api/v1/geo
func GetGeo(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/geo", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.GeoSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		photos, err := query.Geo(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		fc := geojson.NewFeatureCollection()

		bbox := make([]float64, 4)

		bboxMin := func(pos int, val float64) {
			if bbox[pos] == 0.0 || bbox[pos] > val {
				bbox[pos] = val
			}
		}

		bboxMax := func(pos int, val float64) {
			if bbox[pos] == 0.0 || bbox[pos] < val {
				bbox[pos] = val
			}
		}

		for _, p := range photos {
			bboxMin(0, p.Lng())
			bboxMin(1, p.Lat())
			bboxMax(2, p.Lng())
			bboxMax(3, p.Lat())

			feat := geojson.NewPointFeature([]float64{p.Lng(), p.Lat()})
			feat.ID = p.ID
			feat.Properties = gin.H{
				"PhotoUUID":     p.PhotoUUID,
				"PhotoTitle":    p.PhotoTitle,
				"PhotoFavorite": p.PhotoFavorite,
				"FileHash":      p.FileHash,
				"FileWidth":     p.FileWidth,
				"FileHeight":    p.FileHeight,
				"TakenAt":       p.TakenAt,
			}
			fc.AddFeature(feat)
		}

		fc.BoundingBox = bbox

		resp, err := fc.MarshalJSON()

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		c.Data(http.StatusOK, "application/json", resp)
	})
}
