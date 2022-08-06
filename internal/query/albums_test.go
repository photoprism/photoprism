package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAlbumByUID(t *testing.T) {
	t.Run("existing uid", func(t *testing.T) {
		if album, err := AlbumByUID("at9lxuqxpogaaba7"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Christmas 2030", album.AlbumTitle)
		}

		if cached, err := AlbumByUID("at9lxuqxpogaaba7"); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Christmas 2030", cached.AlbumTitle)
		}
	})

	t.Run("not existing uid", func(t *testing.T) {
		album, err := AlbumByUID("3765")
		assert.NotNil(t, album)
		assert.Error(t, err, "record not found")
	})
}

func TestAlbumCoverByUID(t *testing.T) {
	t.Run("existing uid default album", func(t *testing.T) {
		file, err := AlbumCoverByUID("at9lxuqxpogaaba8")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04/bridge2.jpg", file.FileName)
	})

	t.Run("existing uid folder album", func(t *testing.T) {
		file, err := AlbumCoverByUID("at1lxuqipogaaba1")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1990/04/bridge2.jpg", file.FileName)
	})

	t.Run("existing uid empty moment album", func(t *testing.T) {
		file, err := AlbumCoverByUID("at7axuzitogaaiax")

		assert.EqualError(t, err, "no cover found", err)
		assert.Equal(t, "", file.FileName)
	})

	t.Run("not existing uid", func(t *testing.T) {
		file, err := AlbumCoverByUID("3765")
		assert.Error(t, err, "record not found")
		t.Log(file)
	})

	t.Run("existing uid empty month album", func(t *testing.T) {
		file, err := AlbumCoverByUID("at1lxuqipogaabj9")

		assert.EqualError(t, err, "no cover found", err)
		assert.Equal(t, "", file.FileName)
	})
}

func TestUpdateAlbumDates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := UpdateAlbumDates(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestUpdateMissingAlbumEntries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := UpdateMissingAlbumEntries(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAlbumEntryFound(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := AlbumEntryFound("pt9jtxrexxvl0yh0"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGetAlbums(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r, err := Albums(0, 3)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 3, len(r))
	})
}
