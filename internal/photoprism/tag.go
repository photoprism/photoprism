package photoprism

import (
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"strings"
)

type Tag struct {
	gorm.Model
	TagLabel string `gorm:"type:varchar(100);unique_index"`
	TagSlug  string `gorm:"type:varchar(100);unique_index"`
}

func NewTag(label string) *Tag {
	if label == "" {
		label = "unknown"
	}

	tagLabel := strings.ToLower(label)

	tagSlug := slug.MakeLang(tagLabel, "en")

	result := &Tag{
		TagLabel: tagLabel,
		TagSlug:  tagSlug,
	}

	return result
}

func (t *Tag) FirstOrCreate(db *gorm.DB) *Tag {
	db.FirstOrCreate(t, "tag_label = ?", t.TagLabel)

	return t
}
