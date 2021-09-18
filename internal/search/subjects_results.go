package search

// Subject represents a subject search result.
type Subject struct {
	SubjUID      string `json:"UID"`
	MarkerUID    string `json:"MarkerUID"`
	MarkerSrc    string `json:"MarkerSrc,omitempty"`
	SubjType     string `json:"Type"`
	SubjSlug     string `json:"Slug"`
	SubjName     string `json:"Name"`
	SubjAlias    string `json:"Alias"`
	SubjFavorite bool   `json:"Favorite"`
	SubjPrivate  bool   `json:"Private"`
	SubjExcluded bool   `json:"Excluded"`
	FileCount    int    `json:"FileCount"`
	FileHash     string `json:"FileHash"`
	CropArea     string `json:"CropArea"`
}

// SubjectResults represents subject search results.
type SubjectResults []Subject
