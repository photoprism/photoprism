package search

import (
	"time"
)

// Label represents a label search result.
type Label struct {
	ID               uint      `json:"ID"`
	LabelUID         string    `json:"UID"`
	Thumb            string    `json:"Thumb"`
	ThumbSrc         string    `json:"ThumbSrc,omitempty"`
	LabelSlug        string    `json:"Slug"`
	CustomSlug       string    `json:"CustomSlug"`
	LabelName        string    `json:"Name"`
	LabelPriority    int       `json:"Priority"`
	LabelFavorite    bool      `json:"Favorite"`
	LabelDescription string    `json:"Description"`
	LabelNotes       string    `json:"Notes"`
	PhotoCount       int       `json:"PhotoCount"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
	DeletedAt        time.Time `json:"DeletedAt,omitempty"`
}
