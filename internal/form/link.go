package form

// Link represents a link sharing form.
type Link struct {
	ShareSlug   string `json:"Slug"`
	LinkToken   string `json:"Token"`
	LinkExpires int    `json:"Expires"`
	MaxViews    uint   `json:"MaxViews"` // Reserved, should be 0.
	CanComment  bool   `json:"CanComment"`
	CanEdit     bool   `json:"CanEdit"`
}
