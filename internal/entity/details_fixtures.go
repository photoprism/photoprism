package entity

import "github.com/photoprism/photoprism/pkg/txt/report"

type DetailsMap map[string]Details

func (m DetailsMap) Get(name string, photoId uint) Details {
	if result, ok := m[name]; ok {
		result.PhotoID = photoId
		return result
	}

	return Details{PhotoID: photoId}
}

func (m DetailsMap) Pointer(name string, photoId uint) *Details {
	if result, ok := m[name]; ok {
		result.PhotoID = photoId
		return &result
	}

	return &Details{PhotoID: photoId}
}

var DetailsFixtures = DetailsMap{
	"lake": {
		PhotoID:      1000000,
		Keywords:     "nature, frog",
		Notes:        "notes",
		Subject:      "Lake",
		Artist:       "Hans",
		Copyright:    "copy",
		License:      "MIT",
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		KeywordsSrc:  "meta",
		NotesSrc:     "manual",
		SubjectSrc:   "meta",
		ArtistSrc:    "meta",
		CopyrightSrc: "manual",
		LicenseSrc:   "manual",
	},
	"non-photographic": {
		PhotoID:      1000001,
		Keywords:     "screenshot, info",
		Notes:        "notes",
		Subject:      "Non Photographic",
		Artist:       "Hans",
		Copyright:    "copy",
		License:      "MIT",
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		KeywordsSrc:  "",
		NotesSrc:     "",
		SubjectSrc:   "meta",
		ArtistSrc:    "meta",
		CopyrightSrc: "manual",
		LicenseSrc:   "manual",
	},
	"bridge": {
		PhotoID:      1000003,
		Keywords:     "bridge, nature",
		Notes:        "Some Notes!@#$",
		Subject:      "Bridge",
		Artist:       "Jens Mander",
		Copyright:    "Copyright 2020",
		License:      report.NotAssigned,
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		KeywordsSrc:  "meta",
		NotesSrc:     "manual",
		SubjectSrc:   "meta",
		ArtistSrc:    "meta",
		CopyrightSrc: "manual",
		LicenseSrc:   "manual",
	},
}
