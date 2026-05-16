package meta

import (
	"strconv"
	"strings"
)

// FaceRegion represents an XMP face area with normalized coordinates.
type FaceRegion struct {
	Name string
	Type string
	X    float32
	Y    float32
	W    float32
	H    float32
}

// FaceRegions represents a list of XMP face regions.
type FaceRegions []FaceRegion

type xmpRegions struct {
	RegionList struct {
		Bag struct {
			Li []struct {
				Description struct {
					Name string `xml:"http://www.metadataworkinggroup.com/schemas/regions/ Name,attr"`
					Type string `xml:"http://www.metadataworkinggroup.com/schemas/regions/ Type,attr"`
					Area struct {
						X    string `xml:"http://ns.adobe.com/xmp/sType/Area# x,attr"`
						Y    string `xml:"http://ns.adobe.com/xmp/sType/Area# y,attr"`
						W    string `xml:"http://ns.adobe.com/xmp/sType/Area# w,attr"`
						H    string `xml:"http://ns.adobe.com/xmp/sType/Area# h,attr"`
						Unit string `xml:"http://ns.adobe.com/xmp/sType/Area# unit,attr"`
					} `xml:"Area"`
				} `xml:"Description"`
			} `xml:"li"`
		} `xml:"Bag"`
	} `xml:"RegionList"`
}

func parseRegionFloat(s string) float32 {
	if f, err := strconv.ParseFloat(strings.TrimSpace(s), 32); err == nil {
		return float32(f)
	}

	return 0
}

// Left returns the top-left X coordinate used by PhotoPrism markers.
func (r FaceRegion) Left() float32 {
	return r.X - r.W/2
}

// Top returns the top-left Y coordinate used by PhotoPrism markers.
func (r FaceRegion) Top() float32 {
	return r.Y - r.H/2
}

// Valid tests if the region contains a usable normalized face area.
func (r FaceRegion) Valid() bool {
	return strings.EqualFold(r.Type, "face") && r.W > 0 && r.H > 0
}

func (r xmpRegions) FaceRegions() (regions FaceRegions) {
	for _, li := range r.RegionList.Bag.Li {
		d := li.Description

		region := FaceRegion{
			Name: SanitizeString(d.Name),
			Type: SanitizeString(d.Type),
			X:    parseRegionFloat(d.Area.X),
			Y:    parseRegionFloat(d.Area.Y),
			W:    parseRegionFloat(d.Area.W),
			H:    parseRegionFloat(d.Area.H),
		}

		if region.Valid() {
			regions = append(regions, region)
		}
	}

	return regions
}
