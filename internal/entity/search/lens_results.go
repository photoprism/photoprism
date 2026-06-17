package search

import (
	"time"
)

// Lens represents a lens search result.
type Lens struct {
	ID              uint      `json:"ID"`
	LensSlug        string    `json:"Slug"`
	LensName        string    `json:"Name"`
	LensMake        string    `json:"Make"`
	LensModel       string    `json:"Model"`
	LensType        string    `json:"Type"`
	LensDescription string    `json:"Description"`
	LensNotes       string    `json:"Notes"`
	CreatedAt       time.Time `json:"CreatedAt"`
	UpdatedAt       time.Time `json:"UpdatedAt"`
	DeletedAt       time.Time `json:"DeletedAt"`
}
