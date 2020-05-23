package query

import (
	"time"
)

// LabelResult contains found labels
type LabelResult struct {
	// Label
	ID               uint      `json:"ID"`
	LabelUID         string    `json:"UID"`
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
