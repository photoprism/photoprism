package photoprism

import (
	"encoding/json"
	"fmt"
	. "github.com/photoprism/photoprism/internal/models"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

type OpenstreetmapAddress struct {
	Town        string `json:"town"`
	City        string `json:"city"`
	Postcode    string `json:"postcode"`
	County      string `json:"county"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

type OpenstreetmapLocation struct {
	PlaceId     string                `json:"place_id"`
	Lat         string                `json:"lat"`
	Lon         string                `json:"lon"`
	Name        string                `json:"name"`
	Category    string                `json:"category"`
	Type        string                `json:"type"`
	DisplayName string                `json:"display_name"`
	Address     *OpenstreetmapAddress `json:"address"`
}

// See https://wiki.openstreetmap.org/wiki/Nominatim#Reverse_Geocoding
func (m *MediaFile) GetLocation() (*Location, error) {
	if m.location != nil {
		return m.location, nil
	}

	location := &Location{}

	openstreetmapLocation := &OpenstreetmapLocation{
		Address: &OpenstreetmapAddress{},
	}

	if exifData, err := m.GetExifData(); err == nil {
		url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=jsonv2", exifData.Lat, exifData.Long)

		if res, err := http.Get(url); err == nil {
			json.NewDecoder(res.Body).Decode(openstreetmapLocation)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	if id, err := strconv.Atoi(openstreetmapLocation.PlaceId); err == nil && id > 0 {
		location.ID = uint(id)
	} else {
		return nil, errors.New("no location found")
	}

	if openstreetmapLocation.Address.City != "" {
		location.LocCity = openstreetmapLocation.Address.City
	} else {
		location.LocCity = openstreetmapLocation.Address.Town
	}

	if lat, err := strconv.ParseFloat(openstreetmapLocation.Lat, 64); err == nil {
		location.LocLat = lat
	}

	if lon, err := strconv.ParseFloat(openstreetmapLocation.Lon, 64); err == nil {
		location.LocLong = lon
	}

	location.LocName = openstreetmapLocation.Name
	location.LocPostcode = openstreetmapLocation.Address.Postcode
	location.LocCounty = openstreetmapLocation.Address.County
	location.LocState = openstreetmapLocation.Address.State
	location.LocCountry = openstreetmapLocation.Address.Country
	location.LocCountryCode = openstreetmapLocation.Address.CountryCode
	location.LocDisplayName = openstreetmapLocation.DisplayName
	location.LocCategory = openstreetmapLocation.Category

	if openstreetmapLocation.Type != "yes" && openstreetmapLocation.Type != "unclassified" {
		location.LocType = openstreetmapLocation.Type
	}

	m.location = location

	return m.location, nil
}
