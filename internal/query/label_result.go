package query

import (
	"time"
)

// LabelResult contains found labels
type LabelResult struct {
	// Label
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        time.Time
	LabelUUID        string
	LabelSlug        string
	CustomSlug       string
	LabelName        string
	LabelPriority    int
	LabelCount       int
	LabelFavorite    bool
	LabelDescription string
	LabelNotes       string
}
