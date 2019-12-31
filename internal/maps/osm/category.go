package osm

import "fmt"

func (l Location) Category() (result string) {
	key := fmt.Sprintf("%s=%s", l.LocCategory, l.LocType)
	catKey := fmt.Sprintf("%s=*", l.LocCategory)
	typeKey := fmt.Sprintf("*=%s", l.LocType)

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
