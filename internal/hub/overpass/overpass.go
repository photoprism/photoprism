package overpass

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/dsoprea/go-logging"
	"github.com/photoprism/photoprism/pkg/s2"
)

const OverpassStateQuery = `
is_in(%f,%f);
area._[admin_level="4"];
out meta;
`

const OverpassUrl = "https://overpass-api.de/api/interpreter"

// OverpassResponse represents the response from the Overpass API.
type OverpassResponse struct {
	Version   float32           `json:"version"`
	Generator string            `json:"generator"`
	Elements  []OverpassElement `json:"elements"`
}

// OverpassElement represents a generic Overpass element.
type OverpassElement struct {
	ID   int               `json:"id"`
	Type string            `json:"type"`
	Tags map[string]string `json:"tags"`
}

// Name returns the native name of the Overpass element.
func (e OverpassElement) Name() string {
	return e.Tags["name"]
}

// InternationalName returns the international name of the Overpass element.
func (e OverpassElement) InternationalName() string {
	return e.Tags["int_name"]
}

// LocalizedNames returns a mapping of the available localized names (language ISO code -> name).
func (e OverpassElement) LocalizedNames() map[string]string {
	names := make(map[string]string)

	for name, value := range e.Tags {
		if strings.HasPrefix(name, "name:") {
			country_code := name[len("name:"):]
			names[country_code] = value
		}
	}

	return names
}

// AdministrativeLevel returns the administrative level of the Overpass element if available.
func (e OverpassElement) AdministrativeLevel() string {
	return e.Tags["admin_level"]
}

// FindState queries the Overpass API to retrieve the state name in the native language for the given s2 cell.
func FindState(token string) (state string, err error) {
	lat, lng := s2.LatLng(token)
	query := fmt.Sprintf(OverpassStateQuery, lat, lng)

	r, err := queryOverpass(query)
	if err != nil {
		return state, err
	}

	if len(r.Elements) == 0 {
		return state, fmt.Errorf("overpass: token %s does not have state data", token)
	}

	// TODO Should we return the "native" name or the international?
	state = r.Elements[0].Name()

	return state, nil
}

// queryOverpass sends the given query to the Overpass API and unmarshals the response.
func queryOverpass(query string) (result OverpassResponse, err error) {
	reader := strings.NewReader(fmt.Sprintf("data=[out:json];%s", query))
	req, err := http.NewRequest(http.MethodPost, OverpassUrl, reader)

	if err != nil {
		log.Errorf("overpass: %s", err.Error())
		return result, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	r, err := client.Do(req)

	if err != nil {
		log.Errorf("overpass: %s (http request)", err.Error())
		return result, err
	} else if r.StatusCode >= 400 {
		err = fmt.Errorf("overpass: request failed with code %d", r.StatusCode)
		return result, err
	}

	err = json.NewDecoder(r.Body).Decode(&result)

	if err != nil {
		log.Errorf("overpass: %s (decode json)", err.Error())
		return result, err
	}

	return result, nil
}
