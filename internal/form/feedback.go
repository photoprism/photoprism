package form

import "github.com/ulule/deepcopier"

// Feedback represents support requests / customer feedback.
type Feedback struct {
	Category    string `json:"Category"`
	Message     string `json:"Message"`
	UserName    string `json:"UserName"`
	UserEmail   string `json:"UserEmail"`
	UserAgent   string `json:"UserAgent"`
	UserLocales string `json:"UserLocales"`
}

func (f Feedback) Empty() bool {
	return len(f.Category) < 1 || len(f.Message) < 3 || len(f.UserEmail) < 5
}

func NewFeedback(m interface{}) (f Feedback, err error) {
	err = deepcopier.Copy(m).To(&f)

	return f, err
}
