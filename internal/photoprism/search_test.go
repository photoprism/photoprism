package photoprism

import (
	"github.com/photoprism/photoprism/internal/forms"
	"testing"
)

func TestSearch_Photos_Query(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewSearch(conf.OriginalsPath, conf.GetDb())

	var form forms.PhotoSearchForm

	form.Query = "african"
	form.Count = 3
	form.Offset = 0

	photos, err := search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)

	photos, err = search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)

}

func TestSearch_Photos_Camera(t *testing.T) {
	conf := NewTestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewSearch(conf.OriginalsPath, conf.GetDb())

	var form forms.PhotoSearchForm

	form.Query = ""
	form.CameraID = 2
	form.Count = 3
	form.Offset = 0

	photos, err := search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)
}
