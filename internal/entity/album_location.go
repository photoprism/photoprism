package entity

// EstimateCountry updates the album with an estimated country if possible, by taking all photos in the folder into account.
func (m *Album) EstimateCountry() {
	photoCountries := map[string]bool{}

	for _, pa := range m.Photos {
		photo := pa.Photo
		if photo.HasCountry() {
			photoCountries[photo.CountryCode()] = true
		}
	}

	if len(photoCountries) == 1 {
		countryCodes := make([]string, 0, len(photoCountries))

		for cc := range photoCountries {
			countryCodes = append(countryCodes, cc)
		}

		m.AlbumCountry = countryCodes[0]
	}
}
