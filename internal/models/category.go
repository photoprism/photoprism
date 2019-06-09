package models

// Labels can have zero or more categories with the same or a similar meaning
type Category struct {
	LabelID    uint `gorm:"primary_key;auto_increment:false"`
	CategoryID uint `gorm:"primary_key;auto_increment:false"`
	Label      *Label
	Category   *Label
}

func (Category) TableName() string {
	return "categories"
}
