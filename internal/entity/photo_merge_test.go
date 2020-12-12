package entity

import "testing"

func TestPhoto_IdenticalIdentical(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photo := PhotoFixtures.Get("Photo19")

		result, err := photo.Identical(true, true)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("result: %#v", result)
	})
}
