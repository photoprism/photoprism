package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeyword(t *testing.T) {
	t.Run("Cat", func(t *testing.T) {
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

func TestKeyword_TableName(t *testing.T) {
	keyword := &Keyword{}
	assert.Equal(t, "keywords", keyword.TableName())
}

func TestFirstOrCreateKeyword(t *testing.T) {
	keyword := NewKeyword("food")
	result := FirstOrCreateKeyword(keyword)

	if result == nil {
		t.Fatal("result must not be nil")
	}

	if result.Keyword != keyword.Keyword {
		t.Errorf("Keyword should be the same: %s %s", result.Keyword, keyword.Keyword)
	}
}

func TestKeyword_Updates(t *testing.T) {
	t.Run("Success no ID on keyword", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate")

		assert.NoError(t, keyword.Save())
		assert.Equal(t, "keywordbeforeupdate", keyword.Keyword)

		err := keyword.Updates(Keyword{Keyword: "KeywordAfterUpdate", ID: 999})

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "KeywordAfterUpdate", keyword.Keyword)
		assert.Equal(t, uint(0x3e7), keyword.ID)
	})
	t.Run("NilValues", func(t *testing.T) {
		keyword := NewKeyword("noop")
		assert.NoError(t, keyword.Updates(nil))
	})
	t.Run("NilKeyword", func(t *testing.T) {
		var keyword *Keyword
		err := keyword.Updates(Keyword{Keyword: "value"})
		assert.EqualError(t, err, "keyword must not be nil - you may have found a bug")
	})
	t.Run("MissingID", func(t *testing.T) {
		keyword := NewKeyword("missing-id")
		err := keyword.Updates(Keyword{Keyword: "value"})
		assert.EqualError(t, err, "keyword ID must not be empty - you may have found a bug")
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
			assert.ErrorContains(t, err, "keyword ID must not be empty - you may have found a bug")
		} else {
			assert.Fail(t, "error was expected but not set")
		}
		assert.Equal(t, "keywordbeforeupdate4", keyword.Keyword)
		assert.Equal(t, uint(0x0), keyword.ID)
	})
}

func TestKeyword_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		keyword := NewKeyword("KeywordBeforeUpdate3")

		require.NoError(t, keyword.Save())
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
		assert.ErrorContains(t, err, "keyword ID must not be empty - you may have found a bug")
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

func TestFlushCachedKeyword(t *testing.T) {
	t.Run("DeletesCachedEntry", func(t *testing.T) {
		FlushKeywordCache()
		keyword := NewKeyword(fmt.Sprintf("flush-%d", time.Now().UnixNano()))
		require.NoError(t, keyword.Save())

		keywordCache.SetDefault(keyword.Keyword, keyword)

		FlushCachedKeyword(keyword)

		_, found := keywordCache.Get(keyword.Keyword)
		assert.False(t, found)
	})
	t.Run("NilKeyword", func(t *testing.T) {
		FlushCachedKeyword(nil)
	})
}

func TestKeyword_Create(t *testing.T) {
	FlushKeywordCache()
	keyword := NewKeyword(fmt.Sprintf("keyword-create-%d", time.Now().UnixNano()))
	require.NoError(t, keyword.Create())
	assert.True(t, keyword.HasID())
	var fetched Keyword
	require.NoError(t, Db().First(&fetched, keyword.ID).Error)
	assert.Equal(t, keyword.Keyword, fetched.Keyword)
}

func TestKeyword_HasID(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var keyword *Keyword
		assert.False(t, keyword.HasID())
	})
	t.Run("Unsaved", func(t *testing.T) {
		keyword := NewKeyword("unsaved")
		assert.False(t, keyword.HasID())
	})
	t.Run("Saved", func(t *testing.T) {
		keyword := NewKeyword(fmt.Sprintf("keyword-hasid-%d", time.Now().UnixNano()))
		require.NoError(t, keyword.Save())
		assert.True(t, keyword.HasID())
	})
}

func TestFirstOrCreateKeyword_Cached(t *testing.T) {
	name := fmt.Sprintf("keyword-firstorcreate-%d", time.Now().UnixNano())
	keyword := NewKeyword(name)
	result := FirstOrCreateKeyword(keyword)
	require.NotNil(t, result)
	assert.Equal(t, keyword.Keyword, result.Keyword)

	// Second call should return cached/existing entity with same ID
	cached := FirstOrCreateKeyword(NewKeyword(name))
	require.NotNil(t, cached)
	assert.Equal(t, result.ID, cached.ID)
}
