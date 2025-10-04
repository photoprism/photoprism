package entity

// Category links labels to a root label representing the shared meaning.
type Category struct {
	LabelID    uint   `gorm:"primaryKey;autoIncrement:false"`
	CategoryID uint   `gorm:"primaryKey;autoIncrement:false"`
	Label      *Label `gorm:"foreignKey:LabelID"`
	Category   *Label `gorm:"foreignKey:CategoryID"`
}

// TableName returns the entity table name.
func (Category) TableName() string {
	return "categories"
}
