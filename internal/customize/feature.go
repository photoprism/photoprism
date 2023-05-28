package customize

// FeatureSettings represents feature flags, mainly for the Web UI.
type FeatureSettings struct {
	Account   bool `json:"account" yaml:"Account"`
	Albums    bool `json:"albums" yaml:"Albums"`
	Archive   bool `json:"archive" yaml:"Archive"`
	Delete    bool `json:"delete" yaml:"Delete"`
	Download  bool `json:"download" yaml:"Download"`
	Edit      bool `json:"edit" yaml:"Edit"`
	Estimates bool `json:"estimates" yaml:"Estimates"`
	Favorites bool `json:"favorites" yaml:"Favorites"`
	Files     bool `json:"files" yaml:"Files"`
	Folders   bool `json:"folders" yaml:"Folders"`
	Import    bool `json:"import" yaml:"Import"`
	Labels    bool `json:"labels" yaml:"Labels"`
	Library   bool `json:"library" yaml:"Library"`
	Logs      bool `json:"logs" yaml:"Logs"`
	Moments   bool `json:"moments" yaml:"Moments"`
	People    bool `json:"people" yaml:"People"`
	Places    bool `json:"places" yaml:"Places"`
	Private   bool `json:"private" yaml:"Private"`
	Ratings   bool `json:"ratings" yaml:"Ratings"`
	Reactions bool `json:"reactions" yaml:"Reactions"`
	Review    bool `json:"review" yaml:"Review"`
	Search    bool `json:"search" yaml:"Search"`
	Services  bool `json:"services" yaml:"Services"`
	Settings  bool `json:"settings" yaml:"Settings"`
	Share     bool `json:"share" yaml:"Share"`
	Upload    bool `json:"upload" yaml:"Upload"`
	Videos    bool `json:"videos" yaml:"Videos"`
}
