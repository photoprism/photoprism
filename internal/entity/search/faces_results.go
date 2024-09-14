package search

import (
	"time"
)

// Face represents a face search result.
type Face struct {
	ID              string     `json:"ID"`
	FaceSrc         string     `json:"Src"`
	FaceHidden      bool       `json:"Hidden"`
	FaceDist        float64    `json:"FaceDist,omitempty"`
	SubjUID         string     `json:"SubjUID"`
	SubjSrc         string     `json:"SubjSrc,omitempty"`
	FileUID         string     `json:"FileUID,omitempty"`
	MarkerUID       string     `json:"MarkerUID,omitempty"`
	Samples         int        `json:"Samples"`
	SampleRadius    float64    `json:"SampleRadius"`
	Collisions      int        `json:"Collisions"`
	CollisionRadius float64    `json:"CollisionRadius"`
	MarkerName      string     `json:"Name"`
	Size            int        `json:"Size,omitempty"`
	Score           int        `json:"Score,omitempty"`
	MarkerReview    bool       `json:"Review"`
	MarkerInvalid   bool       `json:"Invalid"`
	Thumb           string     `json:"Thumb"`
	MatchedAt       *time.Time `json:"MatchedAt" yaml:"MatchedAt,omitempty"`
	CreatedAt       time.Time  `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt       time.Time  `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
}

// FaceResults represents face search results.
type FaceResults []Face
