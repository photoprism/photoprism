package search

import (
	"time"
)

// Album represents an album search result.
type Album struct {
	ID               uint      `json:"-"`
	AlbumUID         string    `json:"UID"`
	ParentUID        string    `json:"ParentUID"`
	Thumb            string    `json:"Thumb"`
	ThumbSrc         string    `json:"ThumbSrc,omitempty"`
	AlbumSlug        string    `json:"Slug"`
	AlbumType        string    `json:"Type"`
	AlbumTitle       string    `json:"Title"`
	AlbumLocation    string    `json:"Location"`
	AlbumCategory    string    `json:"Category"`
	AlbumCaption     string    `json:"Caption"`
	AlbumDescription string    `json:"Description"`
	AlbumNotes       string    `json:"Notes"`
	AlbumFilter      string    `json:"Filter"`
	AlbumOrder       string    `json:"Order"`
	AlbumTemplate    string    `json:"Template"`
	AlbumPath        string    `json:"Path"`
	AlbumState       string    `json:"State"`
	AlbumCountry     string    `json:"Country"`
	AlbumYear        int       `json:"Year"`
	AlbumMonth       int       `json:"Month"`
	AlbumDay         int       `json:"Day"`
	AlbumFavorite    bool      `json:"Favorite"`
	AlbumPrivate     bool      `json:"Private"`
	PhotoCount       int       `json:"PhotoCount"`
	LinkCount        int       `json:"LinkCount"`
	CreatedAt        time.Time `json:"CreatedAt"`
	UpdatedAt        time.Time `json:"UpdatedAt"`
	DeletedAt        time.Time `json:"DeletedAt,omitempty"`
}

type AlbumResults []Album
