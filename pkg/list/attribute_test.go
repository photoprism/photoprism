package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFlag(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, ParseKeyValue(""))
	})
	t.Run("Default", func(t *testing.T) {
		f := ParseKeyValue("foo")
		assert.Equal(t, "foo", f.Key)
		assert.Equal(t, "true", f.Value)
	})
	t.Run("True", func(t *testing.T) {
		f := ParseKeyValue("feature:true")
		assert.Equal(t, "feature", f.Key)
		assert.Equal(t, "true", f.Value)
	})
	t.Run("False", func(t *testing.T) {
		f := ParseKeyValue("feature:false")
		assert.Equal(t, "feature", f.Key)
		assert.Equal(t, "false", f.Value)
	})
	t.Run("EmptyValue", func(t *testing.T) {
		f := ParseKeyValue("feature:")
		assert.Equal(t, "feature", f.Key)
		assert.Equal(t, "true", f.Value)
	})
	t.Run("StringValue", func(t *testing.T) {
		f := ParseKeyValue("feature:string")
		assert.Equal(t, "feature", f.Key)
		assert.Equal(t, "string", f.Value)
	})
	t.Run("WhitespaceBetween", func(t *testing.T) {
		f := ParseKeyValue("feature :  string")
		assert.Equal(t, "feature", f.Key)
		assert.Equal(t, "string", f.Value)
	})
	t.Run("WhitespacePadding", func(t *testing.T) {
		f := ParseKeyValue(" feature:string   ")
		assert.Equal(t, "feature", f.Key)
		assert.Equal(t, "string", f.Value)
	})
}

func TestFlag_String(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "foo", ParseKeyValue("foo").String())
	})
	t.Run("True", func(t *testing.T) {
		assert.Equal(t, "feature", ParseKeyValue("feature:true").String())
	})
	t.Run("False", func(t *testing.T) {
		assert.Equal(t, "feature:false", ParseKeyValue("feature:false").String())
	})
	t.Run("EmptyValue", func(t *testing.T) {
		assert.Equal(t, "feature", ParseKeyValue("feature:").String())
	})
	t.Run("StringValue", func(t *testing.T) {
		assert.Equal(t, "feature:string", ParseKeyValue("feature:string").String())
	})
	t.Run("WhitespaceBetween", func(t *testing.T) {
		assert.Equal(t, "feature:string", ParseKeyValue("feature :  string").String())
	})
	t.Run("WhitespacePadding", func(t *testing.T) {
		assert.Equal(t, "*featureq62:String!!#$^&*(", ParseKeyValue(" ^&^&(&*&)feature!q62:String!!#$^&*(   ").String())
	})
	t.Run("SpecialChars", func(t *testing.T) {
		assert.Equal(t, "feature:String!!#$^&*(", ParseKeyValue(" feature:String!!#$^&*(  ").String())
	})
}
