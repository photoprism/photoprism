package search

import (
	"time"
)

// Label represents a label search result.
type Label struct {
	ID               uint      `json:"ID"`
	LabelUID         string    `json:"UID"`
	LabelSlug        string    `json:"Slug"`
	CustomSlug       string    `json:"CustomSlug"`
	LabelName        string    `json:"Name"`
	LabelFavorite    bool      `json:"Favorite"`
	LabelPriority    int       `json:"Priority"`
	LabelNSFW        bool      `json:"NSFW,omitempty"`
	LabelDescription string    `json:"Description"`
	LabelNotes       string    `json:"Notes"`
	PhotoCount       int       `json:"PhotoCount"`
	Thumb            string    `json:"Thumb"`
	ThumbSrc         string    `json:"ThumbSrc,omitempty"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
	DeletedAt        time.Time `json:"DeletedAt,omitempty"`
}
