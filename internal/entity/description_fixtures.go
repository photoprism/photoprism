package entity

type DescriptionMap map[string]Description

func (m DescriptionMap) Get(name string, photoId uint) Description {
	if result, ok := m[name]; ok {
		result.PhotoID = photoId
		return result
	}

	return Description{PhotoID: photoId}
}

func (m DescriptionMap) Pointer(name string, photoId uint) *Description {
	if result, ok := m[name]; ok {
		result.PhotoID = photoId
		return &result
	}

	return &Description{PhotoID: photoId}
}

var DescriptionFixtures = DescriptionMap{
	"lake": {
		PhotoID:          1000000,
		PhotoDescription: "photo description lake",
		PhotoKeywords:    "nature, frog",
		PhotoNotes:       "notes",
		PhotoSubject:     "Lake",
		PhotoArtist:      "Hans",
		PhotoCopyright:   "copy",
		PhotoLicense:     "MIT",
	},
	"blacklist": {
		PhotoID:          1000001,
		PhotoDescription: "photo description blacklist",
		PhotoKeywords:    "screenshot, info",
		PhotoNotes:       "notes",
		PhotoSubject:     "Blacklist",
		PhotoArtist:      "Hans",
		PhotoCopyright:   "copy",
		PhotoLicense:     "MIT",
	},
}
