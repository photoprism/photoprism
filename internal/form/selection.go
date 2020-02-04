package form

type Selection struct {
	Photos []string `json:"photos"`
	Albums []string `json:"albums"`
	Labels []string `json:"labels"`
}

func (f Selection) Empty() bool {
	if len(f.Photos) > 0 || len(f.Albums) > 0 || len(f.Labels) > 0 {
		return false
	}

	return true
}
