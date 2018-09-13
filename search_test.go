package photoprism

import (
	"github.com/photoprism/photoprism/forms"
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

	photos, total, err := search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)
	t.Logf("Total Count: %d", total)

	photos, total, err = search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)
	t.Logf("Total Count: %d", total)

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

	photos, total, err := search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)
	t.Logf("Total Count: %d", total)
}