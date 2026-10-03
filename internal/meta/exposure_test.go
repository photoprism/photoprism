package meta

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeExposure(t *testing.T) {
	cases := map[string]string{
		"0.000504":              "1/1984",
		"0.000587890681345016":  "1/1701",
		"0.002":                 "1/500",
		"0.125":                 "1/8",
		"0.25":                  "1/4",
		"0.3":                   "0.3",
		"0.4":                   "0.4",
		"0.8":                   "0.8",
		"0.35":                  "0.3",
		"0.250009":              "1/4",
		"1000000":               "1000000",
		"1/500":                 "1/500",
		"10/1250":               "1/125",
		"3/1000":                "1/333",
		"1/3":                   "0.3",
		"2/5":                   "0.4",
		"4/5":                   "0.8",
		"2/3":                   "0.7",
		"1/2.5":                 "0.4",
		"1 / 60":                "1/60",
		"2/1":                   "2",
		"5/2":                   "2.5",
		"1":                     "1",
		"30":                    "30",
		"\t1/60\n":              "1/60",
		"":                      "",
		"0":                     "",
		"0/1":                   "",
		"1/0":                   "",
		"-0.5":                  "",
		"-1/60":                 "",
		"1/3000000000":          "",
		"1e-320":                "",
		"9e307":                 "",
		"1e400":                 "",
		"1/1e400":               "",
		"NaN":                   "",
		"Inf":                   "",
		"a/b":                   "a/b",
		"bulb":                  "bulb",
		strings.Repeat("x", 64): strings.Repeat("x", 64),
		strings.Repeat("x", 65): "",
	}

	t.Run("Values", func(t *testing.T) {
		for in, want := range cases {
			assert.Equal(t, want, normalizeExposure(in), in)
		}
	})
	t.Run("Idempotent", func(t *testing.T) {
		for in := range cases {
			once := normalizeExposure(in)
			assert.Equal(t, once, normalizeExposure(once), in)
		}
	})
}

func TestParsedFloat(t *testing.T) {
	_, err := strconv.ParseFloat("0.5", 64)
	assert.True(t, parsedFloat(err))
	_, err = strconv.ParseFloat("1e400", 64)
	assert.True(t, parsedFloat(err))
	_, err = strconv.ParseFloat("bulb", 64)
	assert.False(t, parsedFloat(err))
}
