package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKeyword(t *testing.T) {
	t.Run("cat", func(t *testing.T) {
		keyword := NewKeyword("cat")
		assert.Equal(t, "cat", keyword.Keyword)
		assert.Equal(t, false, keyword.Skip)
	})
	t.Run("TABle", func(t *testing.T) {
		keyword := NewKeyword("TABle")
		assert.Equal(t, "table", keyword.Keyword)
		assert.Equal(t, false, keyword.Skip)
	})
}

func TestFirstOrCreateKeyword(t *testing.T) {
	keyword := NewKeyword("food")
	result := FirstOrCreateKeyword(keyword)

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Keyword != keyword.Keyword {
		t.Errorf("Keyword should be the same: %s %s", result.Keyword, keyword.Keyword)
	}
}

func TestKeyword_Updates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate")

		assert.Equal(t, "keywordbeforeupdate", keyword.Keyword)

		err := keyword.Updates(Keyword{Keyword: "KeywordAfterUpdate", ID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "KeywordAfterUpdate", keyword.Keyword)
		assert.Equal(t, uint(0x3e7), keyword.ID)
	})
}

func TestKeyword_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate2")
		assert.Equal(t, "keywordbeforeupdate2", keyword.Keyword)

		err := keyword.Update("Keyword", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", keyword.Keyword)

	})
}

func TestKeyword_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		keyword := NewKeyword("KeywordName")

		err := keyword.Save()

		if err != nil {
			t.Fatal(err)
		}
	})
}
