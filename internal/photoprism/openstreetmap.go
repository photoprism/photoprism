package photoprism

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/internal/models"
	"github.com/pkg/errors"
)

type openstreetmapAddress struct {
	HouseNumber string `json:"house_number"`
	Road        string `json:"road"`
	Suburb      string `json:"suburb"`
	Town        string `json:"town"`
	City        string `json:"city"`
	Postcode    string `json:"postcode"`
	County      string `json:"county"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

type openstreetmapLocation struct {
	PlaceID     string                `json:"place_id"`
	Lat         string                `json:"lat"`
	Lon         string                `json:"lon"`
	Name        string                `json:"name"`
	Category    string                `json:"category"`
	Type        string                `json:"type"`
	DisplayName string                `json:"display_name"`
	Address     *openstreetmapAddress `json:"address"`
}

// GetLocation See https://wiki.openstreetmap.org/wiki/Nominatim#Reverse_Geocoding
func (m *MediaFile) GetLocation() (*models.Location, error) {
	if m.location != nil {
		return m.location, nil
	}

	location := &models.Location{}

	openstreetmapLocation := &openstreetmapLocation{
		Address: &openstreetmapAddress{},
	}

	if exifData, err := m.GetExifData(); err == nil {
		if exifData.Lat == 0 && exifData.Long == 0 {
			return nil, errors.New("lat and long are missing in Exif metadata")
		}

		url := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=jsonv2", exifData.Lat, exifData.Long)

		if res, err := http.Get(url); err == nil {
			err = json.NewDecoder(res.Body).Decode(openstreetmapLocation)

			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}

	if id, err := strconv.Atoi(openstreetmapLocation.PlaceID); err == nil && id > 0 {
		location.ID = uint(id)
	} else {
		return nil, errors.New("query returned no result")
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

	location.LocName = strings.Title(openstreetmapLocation.Name)
	location.LocHouseNr = openstreetmapLocation.Address.HouseNumber
	location.LocStreet = openstreetmapLocation.Address.Road
	location.LocSuburb = openstreetmapLocation.Address.Suburb
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
