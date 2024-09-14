package entity

// PhotoInterface represents an abstract Photo entity interface.
type PhotoInterface interface {
	GetID() uint
	HasID() bool
	GetUID() string
	Approve() error
	Restore() error
}

// PhotosInterface represents a Photo slice provider interface.
type PhotosInterface interface {
	UIDs() []string
	Photos() []PhotoInterface
}
