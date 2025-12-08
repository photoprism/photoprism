package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandVars(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		assert.Equal(t, "", ExpandVars("", nil))
	})
	t.Run("DefaultPortalUrl", func(t *testing.T) {
		assert.Equal(t,
			"https://portal.foo.bar.baz",
			ExpandVars(DefaultPortalUrl, Vars{"PHOTOPRISM_CLUSTER_DOMAIN": "foo.bar.baz"}))
	})
	t.Run("UnbracedUppercase", func(t *testing.T) {
		in := "https://portal.$CLUSTER_DOMAIN"
		out := ExpandVars(in, Vars{"CLUSTER_DOMAIN": "example.com"})
		assert.Equal(t, "https://portal.example.com", out)
	})
	t.Run("HyphenKeyWithBraces", func(t *testing.T) {
		in := "https://portal.${cluster-domain}"
		out := ExpandVars(in, Vars{"cluster-domain": "foo.bar"})
		assert.Equal(t, "https://portal.foo.bar", out)
	})
	t.Run("MultipleVariablesMixedForms", func(t *testing.T) {
		in := "https://${cluster-domain}/$CLUSTER_DOMAIN"
		out := ExpandVars(in, Vars{
			"cluster-domain": "foo.bar",
			"CLUSTER_DOMAIN": "baz.qux",
		})
		assert.Equal(t, "https://foo.bar/baz.qux", out)
	})
	t.Run("UnknownVarBecomesEmpty", func(t *testing.T) {
		in := "pre $UNKNOWN post"
		out := ExpandVars(in, nil)
		// $UNKNOWN maps to empty -> double space remains between words.
		assert.Equal(t, "pre  post", out)
	})
	t.Run("TrailingDollarIsLiteral", func(t *testing.T) {
		in := "end$"
		out := ExpandVars(in, nil)
		// A trailing '$' is not followed by a name, so it remains.
		assert.Equal(t, "end$", out)
	})
	t.Run("BadSyntaxMissingRightBrace", func(t *testing.T) {
		in := "pre ${foo"
		out := ExpandVars(in, Vars{"foo": "X"})
		// os.Expand eats the invalid "${" sequence; remaining text stays.
		assert.Equal(t, "pre foo", out)
	})
	t.Run("EmptyBracesAreEaten", func(t *testing.T) {
		in := "a ${} b"
		out := ExpandVars(in, nil)
		// os.Expand treats ${} as bad syntax and removes it entirely.
		assert.Equal(t, "a  b", out)
	})
	t.Run("SpecialVarDollar", func(t *testing.T) {
		in := "cost $$100"
		out := ExpandVars(in, nil)
		// In os.Expand, '$$' is parsed as special var "$" and maps to empty.
		assert.Equal(t, "cost 100", out)
	})
}
