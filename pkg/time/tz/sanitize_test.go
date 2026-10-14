package tz

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// frontendZoneNames returns the IANA names the web UI offers, as listed in the generated fixture.
func frontendZoneNames(t *testing.T) []string {
	t.Helper()

	data, err := os.ReadFile("testdata/time-zones-names.txt")
	require.NoError(t, err)

	names := strings.Fields(string(data))
	require.NotEmpty(t, names)

	return names
}

// frontendOffsetIDs returns the offset IDs the web UI offers, read from its options source.
func frontendOffsetIDs(t *testing.T) []string {
	t.Helper()

	data, err := os.ReadFile("../../../frontend/src/options/options.js")
	require.NoError(t, err)

	block := regexp.MustCompile(`(?s)export const GmtOffsets = \[(.*?)\];`).FindSubmatch(data)
	require.Len(t, block, 2, "GmtOffsets not found")

	var ids []string

	for _, m := range regexp.MustCompile(`ID: "([^"]+)"`).FindAllSubmatch(block[1], -1) {
		ids = append(ids, string(m[1]))
	}

	require.NotEmpty(t, ids)

	return ids
}

func TestSanitize(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		assert.Equal(t, Local, Sanitize("Local"))
		assert.Equal(t, Local, Sanitize("local"))
		assert.Equal(t, UTC, Sanitize("UTC"))
		assert.Equal(t, UTC, Sanitize(" utc "))
		assert.Equal(t, "Europe/Berlin", Sanitize("Europe/Berlin\n"))
		assert.Equal(t, UTC, Sanitize("Etc/UTC"))
		assert.Equal(t, GMT, Sanitize("GMT"))
	})
	t.Run("Offsets", func(t *testing.T) {
		assert.Equal(t, "UTC+2", Sanitize("+02:00"))
		assert.Equal(t, "UTC-5", Sanitize("UTC-05:00"))
		assert.Equal(t, "UTC-10", Sanitize("Etc/GMT+10"))
		assert.Equal(t, "Etc/GMT-14", Sanitize("Etc/GMT-14"))
	})
	t.Run("Names", func(t *testing.T) {
		for _, s := range []string{"Europe/Berlin", "America/Argentina/Buenos_Aires", "America/Port-au-Prince", "Asia/Kathmandu"} {
			assert.Equal(t, s, Sanitize(s))
		}
	})
	t.Run("FrontendZones", func(t *testing.T) {
		for _, s := range frontendZoneNames(t) {
			assert.Equal(t, s, Sanitize(s))
		}
		for _, s := range frontendOffsetIDs(t) {
			assert.Equal(t, s, Sanitize(s))
		}
	})
	t.Run("Dropped", func(t *testing.T) {
		for _, s := range []string{
			"",
			"Europe/Berlin\x00",
			"Europe/\x1bBerlin",
			"Europe/\nBerlin",
			"Europe/Ber lin",
			"Europe//Berlin",
			"/Europe/Berlin",
			"Europe/Berlin/",
			"Europe/Berlín",
			"Europe/Berlin (CET)",
			"./Europe/Berlin",
			"UTC+99",
			"UTC+13",
			"GMT+20",
			"utc+5:30x",
			"Etc/GMT+99",
			"9Europe/Berlin",
			strings.Repeat("A", MaxLen+1),
		} {
			assert.Equal(t, "", Sanitize(s), "%q", s)
		}
	})
}

func TestValidZoneName(t *testing.T) {
	assert.True(t, validZoneName("Europe/Berlin"))
	assert.True(t, validZoneName("Etc/GMT-13"))
	assert.False(t, validZoneName("Etc/GMT-15"))
	assert.False(t, validZoneName("Europe/"))
	assert.False(t, validZoneName("Europe/Ber.lin"))
}

func TestIsLetter(t *testing.T) {
	assert.True(t, isLetter('a'))
	assert.True(t, isLetter('Z'))
	assert.False(t, isLetter('1'))
	assert.False(t, isLetter('_'))
}

// TestSanitize_FixtureCurrent checks the fixture against the installed frontend dependency, when present.
func TestSanitize_FixtureCurrent(t *testing.T) {
	data, err := os.ReadFile("../../../node_modules/@vvo/tzdb/time-zones-names.json")

	if err != nil {
		t.Skip("frontend dependencies are not installed")
	}

	var names []string
	require.NoError(t, json.Unmarshal(data, &names))
	assert.Equal(t, names, frontendZoneNames(t), "regenerate testdata/time-zones-names.txt")
}
