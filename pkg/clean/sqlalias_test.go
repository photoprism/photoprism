package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSqlAlias(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		for _, s := range []string{"m", "m2", "faces", "_m", "Marker_2", strings.Repeat("a", SqlAliasMax)} {
			assert.Equal(t, s, SqlAlias(s), "%s is a bare identifier", s)
		}
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlAlias(""))
	})
	t.Run("TooLong", func(t *testing.T) {
		assert.Equal(t, "", SqlAlias(strings.Repeat("a", SqlAliasMax+1)))
	})
	t.Run("LeadingDigit", func(t *testing.T) {
		assert.Equal(t, "", SqlAlias("2m"))
	})
	// Rejected whole rather than stripped: "m2 OR 1=1" must not come back as a usable "m2OR11".
	t.Run("Injection", func(t *testing.T) {
		for _, s := range []string{
			"m2 OR 1=1",
			"m2; DROP TABLE markers",
			"m2--",
			"m2'",
			`m2"`,
			"m2`",
			"m2.markers",
			"m2 ",
			" m2",
			"m2\n",
			"m2/*x*/",
			"m2(",
			"markers WHERE 1=1",
		} {
			assert.Equal(t, "", SqlAlias(s), "%q must be rejected", s)
		}
	})
	t.Run("NonASCII", func(t *testing.T) {
		assert.Equal(t, "", SqlAlias("mä"))
		assert.Equal(t, "", SqlAlias("m\x00"))
		assert.Equal(t, "", SqlAlias("m\t"))
	})
}
