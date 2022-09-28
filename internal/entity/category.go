package entity

// Category of labels regroups labels with the same or a similar meaning using a main/root label
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
