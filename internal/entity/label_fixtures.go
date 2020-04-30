package entity

import (
	"time"

	"github.com/jinzhu/gorm"
)

var LabelFixtures = map[string]Label{
	"landscape": {
		ID:               1000000,
		LabelUUID:        "lt9k3pw1wowuy3c2",
		LabelSlug:        "landscape",
		CustomSlug:       "landscape",
		LabelName:        "Landscape",
		LabelPriority:    0,
		LabelFavorite:    true,
		LabelDescription: "",
		LabelNotes:       "",
		LabelCategories:  []*Label{},
		Links:            []Link{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DeletedAt:        nil,
		New:              false,
	},
}

// CreateLabelFixtures inserts known entities into the database for testing.
func CreateLabelFixtures(db *gorm.DB) {
	for _, entity := range LabelFixtures {
		db.Create(&entity)
	}
}
