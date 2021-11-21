package overpass

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOverpassJson(t *testing.T) {
	t.Run("Nepal", func(t *testing.T) {
		overpassJson := `
		{
			"version": 0.6,
			"generator": "Overpass API 0.7.57.1 74a55df1",
			"osm3s": {
				"timestamp_osm_base": "2021-11-21T08:51:55Z",
				"timestamp_areas_base": "2021-11-21T08:27:30Z",
				"copyright": "The data included in this document is from www.openstreetmap.org. The data is made available under ODbL."
			},
			"elements": [{
				"type": "area",
				"id": 3604583291,
				"tags": {
					"ISO3166-2": "NP-4",
					"admin_level": "4",
					"boundary": "historic",
					"int_name": "Eastern Development Region",
					"is_in:continent": "Asia",
					"is_in:country": "Nepal",
					"is_in:country_code": "NP",
					"name": "पुर्वाञ्चल विकास क्षेत्र",
					"name:ar": "المنطقة التنموية الشرقية",
					"name:de": "Entwicklungsregion Ost",
					"name:en": "Eastern Development Region",
					"name:es": "Región de desarrollo Este",
					"name:fr": "Région de développement Est",
					"name:ja": "東部開発区域",
					"name:ne": "पुर्वाञ्चल विकास क्षेत्र",
					"name:pl": "Eastern Development Region",
					"name:ru": "Восточный регион",
					"name:ur": "مشرقی ترقیاتی علاقہ",
					"name:zh": "东部经济发展区",
					"note": "outdated",
					"type": "boundary",
					"wikidata": "Q28576",
					"wikipedia": "en:Eastern Development Region, Nepal"
				}
			}]
		}`

		expectedLocalizedNames := map[string]string{
			"ar": "المنطقة التنموية الشرقية",
			"de": "Entwicklungsregion Ost",
			"en": "Eastern Development Region",
			"es": "Región de desarrollo Este",
			"fr": "Région de développement Est",
			"ja": "東部開発区域",
			"ne": "पुर्वाञ्चल विकास क्षेत्र",
			"pl": "Eastern Development Region",
			"ru": "Восточный регион",
			"ur": "مشرقی ترقیاتی علاقہ",
			"zh": "东部经济发展区",
		}

		var r OverpassResponse

		if err := json.Unmarshal([]byte(overpassJson), &r); err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, r)
		assert.Len(t, r.Elements, 1)

		overpassArea := r.Elements[0]
		assert.Equal(t, "area", overpassArea.Type)
		assert.Equal(t, "4", overpassArea.AdministrativeLevel())
		assert.Equal(t, "पुर्वाञ्चल विकास क्षेत्र", overpassArea.Name())
		assert.Equal(t, "Eastern Development Region", overpassArea.InternationalName())
		assert.Equal(t, expectedLocalizedNames, overpassArea.LocalizedNames())
	})
}

func TestFindState(t *testing.T) {
	t.Run("Khumjung", func(t *testing.T) {
		state, err := FindState("39e9ac0d2c4c")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "पुर्वाञ्चल विकास क्षेत्र", state)
	})
}
