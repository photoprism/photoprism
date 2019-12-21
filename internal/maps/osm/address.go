package osm

type Address struct {
	HouseNumber string `json:"house_number"`
	Road        string `json:"road"`
	Suburb      string `json:"suburb"`
	Town        string `json:"town"`
	Village     string `json:"village"`
	City        string `json:"city"`
	Postcode    string `json:"postcode"`
	County      string `json:"county"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}
