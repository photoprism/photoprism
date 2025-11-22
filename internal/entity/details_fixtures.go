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
	"1000057": {
		PhotoID:      1000057,
		Keywords:     "dog, beach",
		Notes:        "notes",
		Subject:      "Wuff",
		Artist:       "John",
		Copyright:    "My copyright B",
		License:      "N/A",
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		KeywordsSrc:  "meta",
		NotesSrc:     "manual",
		SubjectSrc:   "meta",
		ArtistSrc:    "meta",
		CopyrightSrc: "manual",
		LicenseSrc:   "manual",
	},
	"1000058": {
		PhotoID:      1000058,
		Keywords:     "dog, beach",
		Notes:        "notes",
		Subject:      "Wuff",
		Artist:       "John",
		Copyright:    "My copyright B",
		License:      "N/A",
		CreatedAt:    Now(),
		UpdatedAt:    Now(),
		KeywordsSrc:  "meta",
		NotesSrc:     "manual",
		SubjectSrc:   "meta",
		ArtistSrc:    "meta",
		CopyrightSrc: "manual",
		LicenseSrc:   "manual",
	},
	"1000059": {
		PhotoID:      1000059,
		Keywords:     "christmas, tree",
		Notes:        "notes",
		Subject:      "Tree with lights",
		Artist:       "Santa",
		Copyright:    "My copyright A",
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
}

// CreateDetailsFixtures inserts known entities into the database for testing.
func CreateDetailsFixtures() {
	for _, entity := range DetailsFixtures {
		Db().Create(&entity)
	}
}
