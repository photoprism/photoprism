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
	t.Run("success no ID on keyword", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate")

		assert.Equal(t, "keywordbeforeupdate", keyword.Keyword)

		err := keyword.Updates(Keyword{Keyword: "KeywordAfterUpdate", ID: 999})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "KeywordAfterUpdate", keyword.Keyword)
		assert.Equal(t, uint(0x3e7), keyword.ID)
	})

	t.Run("success ID on keyword", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate3")
		Db().Create(keyword)
		assert.Equal(t, "keywordbeforeupdate3", keyword.Keyword)
		assert.NotEqual(t, 0, keyword.ID)

		err := keyword.Updates(Keyword{Keyword: "KeywordAfterUpdate3"})

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "KeywordAfterUpdate3", keyword.Keyword)
		assert.NotEqual(t, uint(0x3e7), keyword.ID)
	})

	t.Run("failure", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate4")

		assert.Equal(t, "keywordbeforeupdate4", keyword.Keyword)

		err := keyword.Updates(Keyword{Keyword: "KeywordAfterUpdate4"})

		if err != nil {
			assert.Error(t, err)
			assert.ErrorContains(t, err, "id value required but not provided")
		} else {
			assert.Fail(t, "error was expected but not set")
		}
		assert.Equal(t, "keywordbeforeupdate4", keyword.Keyword)
		assert.Equal(t, uint(0x0), keyword.ID)
	})
}

func TestKeyword_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate3")
		assert.Equal(t, "keywordbeforeupdate3", keyword.Keyword)

		keyword.ID = 99966 // Gorm2 requires PK to be set on Model if not using Where clause.
		err := keyword.Update("Keyword", "new-name")

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "new-name", keyword.Keyword)

	})

	t.Run("failure", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate6")
		assert.Equal(t, "keywordbeforeupdate6", keyword.Keyword)

		err := keyword.Update("Keyword", "new-name")
		assert.Error(t, err)
		assert.ErrorContains(t, err, "id value required but not provided")
	})
}

func TestKeyword_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keyword := NewKeyword("KeywordName")

		err := keyword.Save()

		if err != nil {
			t.Fatal(err)
		}
	})
}
