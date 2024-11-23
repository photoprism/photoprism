package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddDuplicate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := AddDuplicate(
			"foobar.jpg",
			RootOriginals,
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd811",
			661851,
			time.Date(2019, 1, 6, 2, 6, 51, 0, time.UTC).Unix(),
		); err != nil {
			t.Fatal(err)
		}

		if err := AddDuplicate(
			"foobar.jpg",
			RootOriginals,
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
			661858,
			time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
		); err != nil {
			t.Fatal(err)
		}

		duplicate := &Duplicate{FileName: "foobar.jpg", FileRoot: RootOriginals}

		if err := duplicate.Find(); err != nil {
			t.Fatal(err)
		} else if duplicate.FileHash != "3cad9168fa6acc5c5c2965ddf6ec465ca42fd818" {
			t.Fatal("file hash should be 3cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		} else if duplicate.ModTime != time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix() {
			t.Fatalf("mod time should be %d", time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix())
		}
	})
	t.Run("error filename empty", func(t *testing.T) {
		err := AddDuplicate(
			"",
			"",
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
			661858,
			time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
		)

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error filehash empty", func(t *testing.T) {
		err := AddDuplicate(
			"foobar.jpg",
			"",
			"",
			661858,
			time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
		)

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error mod time empty", func(t *testing.T) {
		err := AddDuplicate(
			"foobar.jpg",
			"",
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
			661858,
			0,
		)

		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error fileRoot empty", func(t *testing.T) {
		err := AddDuplicate(
			"foobar.jpg",
			"",
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd818",
			661858,
			time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
		)

		if err == nil {
			t.Fatal("error expected")
		}
	})
}

func TestCreateDuplicate(t *testing.T) {
	t.Run("error mod time 0", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "foobar.jpg", FileHash: "12345tghy", FileRoot: RootOriginals, ModTime: 0}
		err := duplicate.Create()
		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error filename empty", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "", FileHash: "12345tghy", FileRoot: RootOriginals, ModTime: time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix()}
		err := duplicate.Create()
		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error filehash empty", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "foobar.jpg", FileHash: "", FileRoot: RootOriginals, ModTime: time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix()}
		err := duplicate.Create()
		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error fileroot empty", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "foobar.jpg", FileHash: "jhg678", FileRoot: "", ModTime: time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix()}
		err := duplicate.Create()
		if err == nil {
			t.Fatal("error expected")
		}
	})
}

func TestSaveDuplicate(t *testing.T) {
	t.Run("error mod time 0", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "foobar.jpg", FileHash: "12345tghy", FileRoot: RootOriginals, ModTime: 0}
		err := duplicate.Save()
		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error filename empty", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "", FileHash: "12345tghy", FileRoot: RootOriginals, ModTime: time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix()}
		err := duplicate.Save()
		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error filehash empty", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "foobar.jpg", FileHash: "", FileRoot: RootOriginals, ModTime: time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix()}
		err := duplicate.Save()
		if err == nil {
			t.Fatal("error expected")
		}
	})
	t.Run("error fileroot empty", func(t *testing.T) {
		duplicate := &Duplicate{FileName: "foobar.jpg", FileHash: "jhg678", FileRoot: "", ModTime: time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix()}
		err := duplicate.Save()
		if err == nil {
			t.Fatal("error expected")
		}
	})
}

func TestDuplicate_Purge(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if err := AddDuplicate(
			"forpurge.jpg",
			RootOriginals,
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd844",
			661851,
			time.Date(2019, 1, 6, 2, 6, 51, 0, time.UTC).Unix(),
		); err != nil {
			t.Fatal(err)
		}

		if err := AddDuplicate(
			"forpurge.jpg",
			RootOriginals,
			"3cad9168fa6acc5c5c2965ddf6ec465ca42fd855",
			661858,
			time.Date(2019, 3, 6, 2, 6, 51, 0, time.UTC).Unix(),
		); err != nil {
			t.Fatal(err)
		}

		duplicate := &Duplicate{FileName: "forpurge.jpg", FileRoot: RootOriginals}
		if err := duplicate.Find(); err != nil {
			t.Fatal(err)
		}
		if err := duplicate.Purge(); err != nil {
			t.Fatal(err)
		}
		if err := duplicate.Find(); err == nil {
			t.Log("Duplicate deleted")
		}
	})
}

func TestPurgeDuplicate(t *testing.T) {
	t.Run("empty filename", func(t *testing.T) {
		assert.Error(t, PurgeDuplicate("", RootOriginals))
	})
	t.Run("empty fileroot", func(t *testing.T) {
		assert.Error(t, PurgeDuplicate("test.jpg", ""))
	})
}
