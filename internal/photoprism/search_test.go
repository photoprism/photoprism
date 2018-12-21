package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/forms"
	"github.com/photoprism/photoprism/internal/test"
)

func TestSearch_Photos_Query(t *testing.T) {
	conf := test.NewConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewSearch(conf.GetOriginalsPath(), conf.Db())

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
	conf := test.NewConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewSearch(conf.GetOriginalsPath(), conf.Db())

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
