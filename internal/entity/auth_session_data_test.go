package entity

import (
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

func TestSessionData_SharedUIDs(t *testing.T) {
	data := SessionData{Shares: []string{"abc123", "def444"},
		Tokens: []string{"5jxf3jfn2k"}}
	assert.Equal(t, "abc123", data.SharedUIDs()[0])
	data2 := SessionData{Shares: []string{},
		Tokens: []string{"5jxf3jfn2k"}}
	assert.Equal(t, "fs6sg6bw45bn0004", data2.SharedUIDs()[0])

}
