package customize

// SearchSettings represents search UI preferences.
type SearchSettings struct {
	BatchSize int  `json:"batchSize" yaml:"BatchSize"`
	ListView  bool `json:"listView" yaml:"ListView"`
}
