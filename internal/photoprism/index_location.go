package photoprism

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/txt"
)

func IndexLocation(db *gorm.DB, conf *config.Config, location *entity.Location, photo *entity.Photo, labels classify.Labels, fileChanged bool, o IndexOptions) ([]string, classify.Labels) {
	location.Lock()
	defer location.Unlock()

	var keywords []string

	err := location.Find(db, conf.GeoCodingApi())

	if err == nil {
		if location.Place.New {
			event.Publish("count.places", event.Data{
				"count": 1,
			})
		}

		photo.Location = location
		photo.LocationID = location.ID
		photo.Place = location.Place
		photo.PlaceID = location.PlaceID
		photo.LocationEstimated = false

		country := entity.NewCountry(location.CountryCode(), location.CountryName()).FirstOrCreate(db)

		if country.New {
			event.Publish("count.countries", event.Data{
				"count": 1,
			})
		}

		locCategory := location.Category()
		keywords = append(keywords, location.Keywords()...)

		// Append category from reverse location lookup
		if locCategory != "" {
			labels = append(labels, classify.LocationLabel(locCategory, 0, -1))
		}

		if (fileChanged || o.UpdateTitle) && !photo.ModifiedTitle {
			if title := labels.Title(location.Name()); title != "" { // TODO: User defined title format
				log.Infof("index: using label \"%s\" to create photo title", title)
				if location.NoCity() || location.LongCity() || location.CityContains(title) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", txt.Title(title), location.CountryName(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", txt.Title(title), location.City(), photo.TakenAt.Format("2006"))
				}
			} else if location.Name() != "" && location.City() != "" {
				if len(location.Name()) > 45 {
					photo.PhotoTitle = txt.Title(location.Name())
				} else if len(location.Name()) > 20 || len(location.City()) > 16 || strings.Contains(location.Name(), location.City()) {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.Name(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.Name(), location.City(), photo.TakenAt.Format("2006"))
				}
			} else if location.City() != "" && location.CountryName() != "" {
				if len(location.City()) > 20 {
					photo.PhotoTitle = fmt.Sprintf("%s / %s", location.City(), photo.TakenAt.Format("2006"))
				} else {
					photo.PhotoTitle = fmt.Sprintf("%s / %s / %s", location.City(), location.CountryName(), photo.TakenAt.Format("2006"))
				}
			}

			if photo.NoTitle() {
				log.Warn("index: could not set photo title based on location or labels")
			} else {
				log.Infof("index: new photo title is \"%s\"", photo.PhotoTitle)
			}
		}
	} else {
		log.Warn(err)

		photo.Place = entity.UnknownPlace
		photo.PlaceID = entity.UnknownPlace.ID
	}

	if !photo.ModifiedLocation || photo.PhotoCountry == "" || photo.PhotoCountry == "zz" {
		photo.PhotoCountry = photo.Place.LocCountry
	}

	return keywords, labels
}

func (ind *Index) estimateLocation(photo *entity.Photo) {
	var recentPhoto entity.Photo

	if result := ind.db.Unscoped().Order(gorm.Expr("ABS(DATEDIFF(taken_at, ?)) ASC", photo.TakenAt)).Preload("Place").First(&recentPhoto); result.Error == nil {
		if recentPhoto.HasPlace() {
			photo.Place = recentPhoto.Place
			photo.PhotoCountry = photo.Place.LocCountry
			photo.LocationEstimated = true
			log.Debugf("index: approximate location is \"%s\"", recentPhoto.Place.Label())
		}
	}
}
