package api

import (
	"net/http"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/form"

	geojson "github.com/paulmach/go.geojson"
)

// GET /api/v1/geo
func GetGeo(router *gin.RouterGroup) {
	router.GET("/geo", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourcePhotos, acl.ActionSearch)

		if s.Invalid() {
			AbortUnauthorized(c)
			return
		}

		var f form.GeoSearch

		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			AbortBadRequest(c)
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

			props := gin.H{
				"UID":     p.PhotoUID,
				"Hash":    p.FileHash,
				"Width":   p.FileWidth,
				"Height":  p.FileHeight,
				"TakenAt": p.TakenAt,
				"Title":   p.PhotoTitle,
			}

			if p.PhotoDescription != "" {
				props["Description"] = p.PhotoDescription
			}

			if p.PhotoType != entity.TypeImage && p.PhotoType != entity.TypeDefault {
				props["Type"] = p.PhotoType
			}

			if p.PhotoFavorite {
				props["Favorite"] = true
			}

			feat := geojson.NewPointFeature([]float64{p.Lng(), p.Lat()})
			feat.ID = p.ID
			feat.Properties = props
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
