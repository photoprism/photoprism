package models

import (
	"github.com/jinzhu/gorm"
)

type Location struct {
	gorm.Model
	LocDisplayName string
	LocLat         float64
	LocLong        float64
	LocCategory    string
	LocType        string
	LocName        string
	LocCity        string
	LocPostcode    string
	LocCounty      string
	LocState       string
	LocCountry     string
	LocCountryCode string
	LocFavorite    bool
}
