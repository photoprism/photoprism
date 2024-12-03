package entity

// Category of labels regroups labels with the same or a similar meaning using a main/root label
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
