package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/forms"
)

func TestSearch_Photos_Query(t *testing.T) {
	conf := config.TestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewSearch(conf.OriginalsPath(), conf.Db())

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
	conf := config.TestConfig()

	conf.CreateDirectories()

	conf.InitializeTestData(t)

	search := NewSearch(conf.OriginalsPath(), conf.Db())

	var form forms.PhotoSearchForm

	form.Query = ""
	form.Camera = 2
	form.Count = 3
	form.Offset = 0

	photos, err := search.Photos(form)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(photos)
}
