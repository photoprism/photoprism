package osm

import "fmt"

func (o Location) Category() (result string) {
	key := fmt.Sprintf("%s=%s", o.LocCategory, o.LocType)
	catKey := fmt.Sprintf("%s=*", o.LocCategory)
	typeKey := fmt.Sprintf("*=%s", o.LocType)

	if result, ok := osmCategories[key]; ok {
		return result
	} else if result, ok := osmCategories[catKey]; ok {
		return result
	} else if result, ok := osmCategories[typeKey]; ok {
		return result
	}

	// log.Debugf("osm: no label found for %s", key)

	return ""
}
