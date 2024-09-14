package customize

// StackSettings represents settings for files that belong to the same photo.
type StackSettings struct {
	UUID bool `json:"uuid" yaml:"UUID"`
	Meta bool `json:"meta" yaml:"Meta"`
	Name bool `json:"name" yaml:"Name"`
}
