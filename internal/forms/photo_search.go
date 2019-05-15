package forms

import (
	"time"
)

// Query parameters for GET /api/v1/photos
type PhotoSearchForm struct {
	Query         string    `form:"q"`
	Location      bool      `form:"location"`
	Tags          string    `form:"tags"`
	Cat           string    `form:"cat"`
	Country       string    `form:"country"`
	CameraID      int       `form:"camera"`
	Order         string    `form:"order"`
	Count         int       `form:"count" binding:"required"`
	Offset        int       `form:"offset"`
	Before        time.Time `form:"before" time_format:"2006-01-02"`
	After         time.Time `form:"after" time_format:"2006-01-02"`
	FavoritesOnly bool      `form:"favorites"`
}
