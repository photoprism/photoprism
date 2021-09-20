package search

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

// Face represents a face search result.
type Face struct {
	ID              string         `json:"ID"`
	FaceSrc         string         `json:"Src"`
	FaceHidden      bool           `json:"Hidden"`
	SubjUID         string         `json:"SubjUID"`
	Samples         int            `json:"Samples"`
	SampleRadius    float64        `json:"SampleRadius"`
	Collisions      int            `json:"Collisions"`
	CollisionRadius float64        `json:"CollisionRadius"`
	Marker          *entity.Marker `json:"Marker,omitempty"`
	MatchedAt       *time.Time     `json:"MatchedAt" yaml:"MatchedAt,omitempty"`
	CreatedAt       time.Time      `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt       time.Time      `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
}

// FaceResults represents face search results.
type FaceResults []Face
