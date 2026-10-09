// Package sample holds one site of each pattern check-gorm-v1 counts, and code it must not count.
package sample

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Item has a DeletedAt field of type *time.Time.
type Item struct {
	ID        uint
	Visible   bool
	DeletedAt *time.Time
}

// Archived has a DeletedAt field of another type.
type Archived struct {
	DeletedAt time.Time
}

// Sites uses each GORM v1 API the check counts once.
func Sites(db *gorm.DB, scope *gorm.Scope) bool {
	var n int
	var n64 int64

	db.Model(&Item{}).Count(&n)
	db.Model(&Item{}).Count(&n64)
	_ = db.Table("items").Select("id").QueryExpr()
	_ = db.Table("items").Select("id").SubQuery()
	_ = db.Dialect().GetName()
	_ = db.Where("photo_private = 0 AND file_missing=1").Where("photo_private = 10 OR xfile_missing = 1").Error

	return db.First(&Item{}).RecordNotFound() || Item{}.DeletedAt != nil
}

// Indirect uses the same APIs through method expressions, method values, package functions and quoted names.
func Indirect(db *gorm.DB, err error) bool {
	var n uint
	count := db.Count

	(*gorm.DB).Count(db, &n)
	count(&n)
	_ = db.NewScope(&Item{})
	_ = db.Where("`photo_private` = 1 AND \"photos\".\"file_missing\"<>0").Error

	return gorm.IsRecordNotFoundError(err)
}

// Other uses names the check must not count.
func Other(a Archived) time.Time {
	return a.DeletedAt
}
