package osm

import "fmt"

var locationLabels = map[string]string{
	"bay":                   "bay",
	"art":                   "art exhibition",
	"fire station":          "fire station",
	"hairdresser":           "hairdresser",
	"cape":                  "cape",
	"coastline":             "coastline",
	"cliff":                 "cliff",
	"wetland":               "wetland",
	"nature reserve":        "nature reserve",
	"natural=beach":         "beach",
	"amenity=cafe":          "cafe",
	"amenity=internet_cafe": "cafe",
	"ice cream":             "ice cream parlor",
	"bistro":                "restaurant",
	"restaurant":            "restaurant",
	"ship":                  "ship",
	"wholesale":             "shop",
	"food":                  "shop",
	"supermarket":           "supermarket",
	"florist":               "florist",
	"pharmacy":              "pharmacy",
	"seafood":               "seafood",
	"clothes":               "clothing store",
	"residential":           "residential area",
	"museum":                "museum",
	"castle":                "castle",
	"aeroway=*":             "airport",
	"ferry terminal":        "harbor",
	"bridge":                "bridge",
	"university":            "university",
	"mall":                  "mall",
	"marina":                "marina",
	"garden":                "garden",
	"pedestrian":            "shopping area",
	"bunker":                "bunker",
	"viewpoint":             "viewpoint",
	"train station":         "train station",
	"farm":                  "farm",
	"highway=secondary":     "highway",
}

func (o *Location) Label() (result string) {
	key := fmt.Sprintf("%s=%s", o.Category, o.Type)
	catKey := fmt.Sprintf("%s=*", o.Category)
	typeKey := fmt.Sprintf("*=%s", o.Type)

	if result, ok := locationLabels[key]; ok {
		return result
	} else if result, ok := locationLabels[catKey]; ok {
		return result
	} else if result, ok := locationLabels[typeKey]; ok {
		return result
	}

	log.Debugf("osm: no label found for %s", key)

	return ""
}
