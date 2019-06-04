package models

// Labels can have zero or more synonyms with the same or a similar meaning
type Synonym struct {
	LabelID   uint `gorm:"primary_key;auto_increment:false"`
	SynonymID uint `gorm:"primary_key;auto_increment:false"`
	Label     *Label
	Synonym   *Label
}

func (Synonym) TableName() string {
	return "synonyms"
}
