package config

// FeatureSettings represents feature flags, mainly for the Web UI.
type FeatureSettings struct {
	Upload    bool `json:"upload" yaml:"Upload"`
	Download  bool `json:"download" yaml:"Download"`
	Private   bool `json:"private" yaml:"Private"`
	Review    bool `json:"review" yaml:"Review"`
	Files     bool `json:"files" yaml:"Files"`
	Videos    bool `json:"videos" yaml:"Videos"`
	Folders   bool `json:"folders" yaml:"Folders"`
	Albums    bool `json:"albums" yaml:"Albums"`
	Moments   bool `json:"moments" yaml:"Moments"`
	Estimates bool `json:"estimates" yaml:"Estimates"`
	People    bool `json:"people" yaml:"People"`
	Labels    bool `json:"labels" yaml:"Labels"`
	Places    bool `json:"places" yaml:"Places"`
	Edit      bool `json:"edit" yaml:"Edit"`
	Archive   bool `json:"archive" yaml:"Archive"`
	Delete    bool `json:"delete" yaml:"Delete"`
	Share     bool `json:"share" yaml:"Share"`
	Library   bool `json:"library" yaml:"Library"`
	Import    bool `json:"import" yaml:"Import"`
	Logs      bool `json:"logs" yaml:"Logs"`
}
