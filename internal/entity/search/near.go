package search

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/geo"
	"github.com/photoprism/photoprism/pkg/geo/s2"
)

// nearSQLCreator uses the near from frm.Near and dist from frm.Dist to generate the qs (Query String) and interface of values to allow Gorm to generate the required clause
func nearSQLCreator(near string, dist float64) (qs string, values []interface{}, err error) {
	photos := []Photo{}
	if err := Db().Model(&Photo{}).Where("photo_uid IN (?)", SplitOr(near)).Select("photo_uid, cell_id").Find(&photos).Error; err != nil {
		log.Debugf("search: %s (find nearby)", err)
		return qs, values, ErrNotFound
	}

	if len(photos) == 0 {
		return qs, values, ErrNotFound
	}

	wheres := make([]string, len(photos))
	for item, photo := range photos {
		// Set the S2 Cell ID to search for.
		s2Cell := photo.CellID

		// Set the search distance if unspecified.
		if dist <= 0 {
			dist = geo.DefaultDist
		}

		wheres[item] = "photos.cell_id BETWEEN ? AND ?"
		s2Min, s2Max := s2.PrefixedRange(s2Cell, s2.Level(dist))
		values = append(values, s2Min)
		values = append(values, s2Max)
	}
	return strings.Join(wheres, " OR "), values, nil
}
