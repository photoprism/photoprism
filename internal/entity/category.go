package entity

// Category links labels to a root label representing the shared meaning.
type Category struct {
	LabelID    uint `gorm:"primary_key;auto_increment:false"`
	CategoryID uint `gorm:"primary_key;auto_increment:false"`
	Label      *Label
	Category   *Label
}

// TableName returns the entity table name.
func (Category) TableName() string {
	return "categories"
}
