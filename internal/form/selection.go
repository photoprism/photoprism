package form

type Selection struct {
	Files  []string `json:"files"`
	Photos []string `json:"photos"`
	Albums []string `json:"albums"`
	Labels []string `json:"labels"`
	Places []string `json:"places"`
}

func (f Selection) Empty() bool {
	switch {
	case len(f.Files) > 0:
		return false
	case len(f.Photos) > 0:
		return false
	case len(f.Albums) > 0:
		return false
	case len(f.Labels) > 0:
		return false
	case len(f.Places) > 0:
		return false
	}

	return true
}
