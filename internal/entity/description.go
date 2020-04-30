package entity

// Description stores additional metadata fields for each photo to improve search performance.
type Description struct {
	PhotoID          uint   `gorm:"primary_key;auto_increment:false"`
	PhotoDescription string `gorm:"type:text;" json:"PhotoDescription"`
	PhotoKeywords    string `gorm:"type:text;" json:"PhotoKeywords"`
	PhotoNotes       string `gorm:"type:text;" json:"PhotoNotes"`
	PhotoSubject     string `gorm:"type:varchar(255);" json:"PhotoSubject"`
	PhotoArtist      string `gorm:"type:varchar(255);" json:"PhotoArtist"`
	PhotoCopyright   string `gorm:"type:varchar(255);" json:"PhotoCopyright"`
	PhotoLicense     string `gorm:"type:varchar(255);" json:"PhotoLicense"`
}

// FirstOrCreate returns the matching entity or creates a new one.
func (m *Description) FirstOrCreate() error {
	return Db().FirstOrCreate(m, "photo_id = ?", m.PhotoID).Error
}

// NoDescription checks if the photo has no Description
func (m *Description) NoDescription() bool {
	return m.PhotoDescription == ""
}

// NoKeywords checks if the photo has no Keywords
func (m *Description) NoKeywords() bool {
	return m.PhotoKeywords == ""
}

// NoSubject checks if the photo has no Subject
func (m *Description) NoSubject() bool {
	return m.PhotoSubject == ""
}

// NoNotes checks if the photo has no Notes
func (m *Description) NoNotes() bool {
	return m.PhotoNotes == ""
}

// NoArtist checks if the photo has no Artist
func (m *Description) NoArtist() bool {
	return m.PhotoArtist == ""
}

// NoCopyright checks if the photo has no Copyright
func (m *Description) NoCopyright() bool {
	return m.PhotoCopyright == ""
}
