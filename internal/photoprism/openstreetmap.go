package photoprism

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
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
	PlaceID     uint                  `json:"place_id"`
	Lat         string                `json:"lat"`
	Lon         string                `json:"lon"`
	Name        string                `json:"name"`
	Category    string                `json:"category"`
	Type        string                `json:"type"`
	DisplayName string                `json:"display_name"`
	Address     *openstreetmapAddress `json:"address"`
}

// Location See https://wiki.openstreetmap.org/wiki/Nominatim#Reverse_Geocoding
func (m *MediaFile) Location() (*models.Location, error) {
	if m.location != nil {
		return m.location, nil
	}

	location := &models.Location{}

	openstreetmapLocation := &openstreetmapLocation{
		Address: &openstreetmapAddress{},
	}

	if exifData, err := m.Exif(); err == nil {
		if exifData.Lat == 0 && exifData.Long == 0 {
			return nil, errors.New("no latitude and longitude in image metadata")
		}

		url := fmt.Sprintf(
			"https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=jsonv2&accept-language=en&zoom=18",
			exifData.Lat,
			exifData.Long)

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

	if id := openstreetmapLocation.PlaceID; id > 0 {
		location.ID = id
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

	if len(openstreetmapLocation.Name) > 1 {
		location.LocName = strings.Replace(openstreetmapLocation.Name, " - ", " / ", -1)
		location.LocName = util.Title(strings.TrimSpace(strings.Replace(location.LocName, "_", " ", -1)))
	}

	location.LocHouseNr = strings.TrimSpace(openstreetmapLocation.Address.HouseNumber)
	location.LocStreet = strings.TrimSpace(openstreetmapLocation.Address.Road)
	location.LocSuburb = strings.TrimSpace(openstreetmapLocation.Address.Suburb)
	location.LocPostcode = strings.TrimSpace(openstreetmapLocation.Address.Postcode)
	location.LocCounty = strings.TrimSpace(openstreetmapLocation.Address.County)
	location.LocState = strings.TrimSpace(openstreetmapLocation.Address.State)
	location.LocCountry = strings.TrimSpace(openstreetmapLocation.Address.Country)
	location.LocCountryCode = strings.TrimSpace(openstreetmapLocation.Address.CountryCode)
	location.LocDisplayName = strings.TrimSpace(openstreetmapLocation.DisplayName)

	locationCategory := strings.TrimSpace(strings.Replace(openstreetmapLocation.Category, "_", " ", -1))
	location.LocCategory = locationCategory

	if openstreetmapLocation.Type != "yes" && openstreetmapLocation.Type != "unclassified" {
		locationType := strings.TrimSpace(strings.Replace(openstreetmapLocation.Type, "_", " ", -1))
		location.LocType = locationType
	}

	m.location = location

	return m.location, nil
}
