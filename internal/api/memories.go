package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// MemoryResult represents a single memory (photos from a specific past year).
type MemoryResult struct {
	Year    int         `json:"year"`
	Month   int         `json:"month"`
	Day     int         `json:"day"`
	Title   string      `json:"title"`
	Photos  interface{} `json:"photos"`
	Count   int         `json:"count"`
}

// SearchMemories finds photos from the same day in past years.
//
//	@Summary		finds photos from the same day in past years
//	@Description	Returns photos taken on this day in previous years, grouped by year.
//	@Id				SearchMemories
//	@Tags			Photos
//	@Produce		json
//	@Success		200				{array}	MemoryResult
//	@Failure		400,401,403,500	{object}	i18n.Response
//	@Router			/api/v1/memories [get]
func SearchMemories(router *gin.RouterGroup) {
	router.GET("/memories", func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		if s.Abort(c) {
			return
		}

		conf := get.Config()
		
		// Get current date or use provided date
		var targetDate time.Time
		dateParam := c.Query("date")
		if dateParam != "" {
			parsed, err := time.Parse("2006-01-02", dateParam)
			if err != nil {
				AbortBadRequest(c, err)
				return
			}
			targetDate = parsed
		} else {
			targetDate = time.Now()
		}

		month := targetDate.Month()
		day := targetDate.Day()
		year := targetDate.Year()

		// Search for photos from the same month and day in past years
		searchForm := form.SearchPhotos{
			Month:  month.String(),
			Day:    string(rune(day)),
			Count:  1000, // Get all matching photos
			Offset: 0,
			Order:  "newest",
		}

		// Respect privacy settings
		if conf.Settings().Features.Private {
			searchForm.Public = true
		}

		results, _, err := search.UserPhotos(searchForm, s)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Group photos by year
		memoriesByYear := make(map[int][]interface{})
		for _, photo := range results {
			if photoResult, ok := photo.(search.PhotoResult); ok {
				photoYear := photoResult.Year
				// Only include photos from past years
				if photoYear < year {
					memoriesByYear[photoYear] = append(memoriesByYear[photoYear], photo)
				}
			}
		}

		// Convert to MemoryResult slice
		var memories []MemoryResult
		for y, photos := range memoriesByYear {
			yearsAgo := year - y
			title := ""
			if yearsAgo == 1 {
				title = "1 year ago"
			} else {
				title = i18n.Msg("%d years ago", yearsAgo)
			}

			memories = append(memories, MemoryResult{
				Year:   y,
				Month:  int(month),
				Day:    day,
				Title:  title,
				Photos: photos,
				Count:  len(photos),
			})
		}

		c.JSON(http.StatusOK, memories)
	})
}
