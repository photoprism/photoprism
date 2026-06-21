package entity

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIDs_String(t *testing.T) {
	uid := UIDs{"dghjkfd", "dfgehrih"}
	assert.Equal(t, "dghjkfd,dfgehrih", uid.String())
}

func TestUIDs_Join(t *testing.T) {
	uid := UIDs{"dghjkfd", "dfgehrih"}
	assert.Equal(t, "dghjkfd|dfgehrih", uid.Join("|"))
}

func TestData_HasShare(t *testing.T) {
	data := SessionData{Shares: []string{"abc123", "def444"}}
	assert.True(t, data.HasShare("def444"))
	assert.False(t, data.HasShare("xxx"))
}

func TestSessionData_RedeemToken(t *testing.T) {
	data := SessionData{Shares: []string{"abc123", "def444"}}
	assert.True(t, data.HasShare("def444"))
	assert.False(t, data.HasShare("as6sg6bxpogaaba8"))
	data.RedeemToken("xxx")
	assert.False(t, data.HasShare("xxx"))
	data.RedeemToken("1jxf3jfn2k")
	assert.True(t, data.HasShare("def444"))
	assert.True(t, data.HasShare("as6sg6bxpogaaba8"))
}

func TestSessionData_SetGroups(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		data := NewSessionData()
		data.SetGroups([]string{"media-acme-admin", "media-acme-viewer"})
		assert.Equal(t, []string{"media-acme-admin", "media-acme-viewer"}, data.Groups)
	})
	t.Run("Empty", func(t *testing.T) {
		data := NewSessionData()
		data.SetGroups([]string{"media-acme-admin"})
		data.SetGroups(nil)
		assert.Nil(t, data.Groups)
	})
	t.Run("HoldsEntraOverageBound", func(t *testing.T) {
		// Entra emits at most 200 groups in a token before signaling overage;
		// a set of that size must be stored without truncation.
		groups := make([]string, 0, 200)
		for i := 0; i < 200; i++ {
			groups = append(groups, "12345678-1234-1234-1234-12345678901"+string(rune('a'+i%26)))
		}
		data := NewSessionData()
		data.SetGroups(groups)
		assert.Len(t, data.Groups, 200)
	})
	t.Run("TruncatesOversizedSet", func(t *testing.T) {
		groups := make([]string, 0, 500)
		for i := 0; i < 500; i++ {
			groups = append(groups, strings.Repeat("a", 30)+string(rune('a'+i%26)))
		}
		data := NewSessionData()
		data.SetGroups(groups)
		assert.NotEmpty(t, data.Groups)
		assert.Less(t, len(data.Groups), len(groups))
		j, err := json.Marshal(data)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(j), 16384, "serialized session data must fit the database column")
	})
}

func TestSessionData_Redacted(t *testing.T) {
	t.Run("StripsGroups", func(t *testing.T) {
		data := &SessionData{
			Tokens: []string{"1jxf3jfn2k"},
			Shares: UIDs{"as6sg6bxpogaaba8"},
			Groups: []string{"media-acme-admin"},
		}
		redacted := data.Redacted()
		assert.Nil(t, redacted.Groups)
		assert.Equal(t, data.Tokens, redacted.Tokens)
		assert.Equal(t, data.Shares, redacted.Shares)
		assert.Equal(t, []string{"media-acme-admin"}, data.Groups, "the original must stay intact")
	})
	t.Run("Nil", func(t *testing.T) {
		var data *SessionData
		assert.Nil(t, data.Redacted())
	})
}

func TestSessionData_SharedUIDs(t *testing.T) {
	data := SessionData{Shares: []string{"abc123", "def444"},
		Tokens: []string{"5jxf3jfn2k"}}
	assert.Equal(t, "abc123", data.SharedUIDs()[0])
	data2 := SessionData{Shares: []string{},
		Tokens: []string{"5jxf3jfn2k"}}
	assert.Equal(t, "fs6sg6bw45bn0004", data2.SharedUIDs()[0])

}
