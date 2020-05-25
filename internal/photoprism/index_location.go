package photoprism

import (
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
)

func (ind *Index) estimateLocation(photo *entity.Photo) {
	var recentPhoto entity.Photo

	if result := ind.db.Unscoped().Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Place").First(&recentPhoto); result.Error == nil {
		if recentPhoto.HasPlace() {
			photo.Place = recentPhoto.Place
			photo.PhotoCountry = photo.Place.LocCountry
			photo.LocSrc = entity.SrcAuto
			log.Debugf("index: approximate location is %s", txt.Quote(recentPhoto.Place.Label()))
		}
	}
}
