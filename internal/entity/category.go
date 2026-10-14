package entity

// Category links labels to a root label representing the shared meaning.
type Category struct {
	LabelID    uint   `gorm:"primaryKey;autoIncrement:false;size:32"`
	CategoryID uint   `gorm:"primaryKey;autoIncrement:false;size:32"`
	Label      *Label `gorm:"-:migration"`
	Category   *Label `gorm:"-:migration"`
}

// TableName returns the entity table name.
func (Category) TableName() string {
	return "categories"
}
