package entity

import (
	"testing"
	"time"
)

func TestAddDuplicate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
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
	t.Run("error", func(t *testing.T) {
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
}
