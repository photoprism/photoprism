package photoprism

import (
	"github.com/photoprism/photoprism/forms"
	"testing"
)

func TestSearch_Photos(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewQuery(conf.OriginalsPath, conf.GetDb())

	var form forms.PhotoSearchForm

	form.Query = "elephant"
	form.Count = 3
	form.Offset = 0

	photos, err := search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)
}
