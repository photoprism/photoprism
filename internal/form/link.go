package form

// Link represents a link sharing form.
type Link struct {
	Password    string `json:"Password"`
	ShareSlug   string `json:"Slug"`
	LinkToken   string `json:"Token"`
	LinkExpires int    `json:"Expires"`
	MaxViews    uint   `json:"MaxViews"`
	CanComment  bool   `json:"CanComment"`
	CanEdit     bool   `json:"CanEdit"`
}
