package photoprism

import (
	"fmt"
	"math"
	"runtime/debug"
	"strconv"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Moments represents a worker that creates albums based on popular locations, dates and labels.
type Moments struct {
	conf *config.Config
}

// NewMoments returns a new Moments worker.
func NewMoments(conf *config.Config) *Moments {
	instance := &Moments{
		conf: conf,
	}

	return instance
}

// MigrateSlug updates deprecated moment slugs if needed.
func (w *Moments) MigrateSlug(m query.Moment, albumType string) {
	if m.Slug() == m.TitleSlug() {
		return
	}

	if a := entity.FindAlbumBySlug(m.TitleSlug(), albumType); a != nil {
		logWarn("moments", a.Update("album_slug", m.Slug()))
	}
}

// Start creates albums based on popular locations, dates and categories.
func (w *Moments) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)\nstack: %s", r, debug.Stack())
			log.Errorf("moments: %s", err)
		}
	}()

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	// Remove duplicate moments.
	if removed, err := query.RemoveDuplicateMoments(); err != nil {
		log.Warnf("moments: %s (remove duplicates)", err)
	} else if removed > 0 {
		log.Infof("moments: removed %s", english.Plural(removed, "duplicate", "duplicates"))
	}

	counts := query.Counts{}
	counts.Refresh()

	indexSize := counts.Photos + counts.Videos

	threshold := 3

	if indexSize > 4 {
		threshold = int(math.Log2(float64(indexSize))) + 1
	}

	log.Debugf("moments: analyzing %d photos and %d videos, with threshold %d", counts.Photos, counts.Videos, threshold)

	if indexSize < threshold {
		log.Debugf("moments: not enough files")

		return nil
	}

	// Important folders.
	if results, err := query.AlbumFolders(1); err != nil {
		log.Errorf("moments: %s", err.Error())
	} else {
		for _, mom := range results {
			f := form.PhotoSearch{
				Path:   mom.Path,
				Public: true,
			}

			if a := entity.FindFolderAlbum(mom.Path); a != nil {
				if a.DeletedAt != nil {
					// Nothing to do.
					log.Tracef("moments: %s was deleted (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				} else if err := a.UpdateFolder(mom.Path, f.Serialize()); err != nil {
					log.Errorf("moments: %s (update folder)", err.Error())
				} else {
					log.Tracef("moments: %s already exists (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			} else if a := entity.NewFolderAlbum(mom.Title(), mom.Path, f.Serialize()); a != nil {
				a.AlbumYear = mom.FolderYear
				a.AlbumMonth = mom.FolderMonth
				a.AlbumDay = mom.FolderDay
				a.AlbumCountry = mom.FolderCountry

				if err := a.Create(); err != nil {
					log.Errorf("moments: %s (create folder)", err)
				} else {
					log.Infof("moments: added %s (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			}
		}
	}

	// All years and months.
	if results, err := query.MomentsTime(1); err != nil {
		log.Errorf("moments: %s", err.Error())
	} else {
		for _, mom := range results {
			if a := entity.FindMonthAlbum(mom.Year, mom.Month); a != nil {
				if err := a.UpdateSlug(mom.Title(), mom.Slug()); err != nil {
					log.Errorf("moments: %s (update slug)", err.Error())
				}

				if !a.Deleted() {
					log.Tracef("moments: %s already exists (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				} else if err := a.Restore(); err != nil {
					log.Errorf("moments: %s (restore month)", err.Error())
				} else {
					log.Infof("moments: %s restored", txt.Quote(a.AlbumTitle))
				}
			} else if a := entity.NewMonthAlbum(mom.Title(), mom.Slug(), mom.Year, mom.Month); a != nil {
				if err := a.Create(); err != nil {
					log.Errorf("moments: %s", err)
				} else {
					log.Infof("moments: added %s (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			}
		}
	}

	// Countries by year.
	if results, err := query.MomentsCountries(threshold); err != nil {
		log.Errorf("moments: %s", err.Error())
	} else {
		for _, mom := range results {
			f := form.PhotoSearch{
				Country: mom.Country,
				Year:    strconv.Itoa(mom.Year),
				Public:  true,
			}

			if a := entity.FindAlbumByAttr(S{mom.Slug(), mom.TitleSlug()}, S{f.Serialize()}, entity.AlbumMoment); a != nil {
				if err := a.UpdateSlug(mom.Title(), mom.Slug()); err != nil {
					log.Errorf("moments: %s (update slug)", err.Error())
				}

				if a.DeletedAt != nil {
					// Nothing to do.
					log.Tracef("moments: %s was deleted (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				} else {
					log.Tracef("moments: %s already exists (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			} else if a := entity.NewMomentsAlbum(mom.Title(), mom.Slug(), f.Serialize()); a != nil {
				a.AlbumYear = mom.Year
				a.AlbumCountry = mom.Country

				if err := a.Create(); err != nil {
					log.Errorf("moments: %s", err)
				} else {
					log.Infof("moments: added %s (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			}
		}
	}

	// States and countries.
	if results, err := query.MomentsStates(1); err != nil {
		log.Errorf("moments: %s", err.Error())
	} else {
		for _, mom := range results {
			f := form.PhotoSearch{
				Country: mom.Country,
				State:   mom.State,
				Public:  true,
			}

			if a := entity.FindAlbumByAttr(S{mom.Slug(), mom.TitleSlug()}, S{f.Serialize()}, entity.AlbumState); a != nil {
				if err := a.UpdateState(mom.Title(), mom.Slug(), mom.State, mom.Country); err != nil {
					log.Errorf("moments: %s (update state)", err.Error())
				}

				if !a.Deleted() {
					log.Tracef("moments: %s already exists (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				} else if err := a.Restore(); err != nil {
					log.Errorf("moments: %s (restore state)", err.Error())
				} else {
					log.Infof("moments: %s restored", txt.Quote(a.AlbumTitle))
				}
			} else if a := entity.NewStateAlbum(mom.Title(), mom.Slug(), f.Serialize()); a != nil {
				a.AlbumLocation = mom.CountryName()
				a.AlbumCountry = mom.Country
				a.AlbumState = mom.State

				if err := a.Create(); err != nil {
					log.Errorf("moments: %s", err)
				} else {
					log.Infof("moments: added %s (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			}
		}
	}

	// Popular labels.
	if results, err := query.MomentsLabels(threshold); err != nil {
		log.Errorf("moments: %s", err.Error())
	} else {
		for _, mom := range results {
			w.MigrateSlug(mom, entity.AlbumMoment)

			f := form.PhotoSearch{
				Label:  mom.Label,
				Public: true,
			}

			if a := entity.FindAlbumByAttr(S{mom.Slug(), mom.TitleSlug()}, S{f.Serialize()}, entity.AlbumMoment); a != nil {
				if err := a.UpdateSlug(mom.Title(), mom.Slug()); err != nil {
					log.Errorf("moments: %s (update slug)", err.Error())
				}

				if a.DeletedAt != nil || f.Serialize() == a.AlbumFilter {
					log.Tracef("moments: %s already exists (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
					continue
				}

				if err := a.Update("AlbumFilter", f.Serialize()); err != nil {
					log.Errorf("moments: %s", err.Error())
				} else {
					log.Debugf("moments: updated %s (%s)", txt.Quote(a.AlbumTitle), f.Serialize())
				}
			} else if a := entity.NewMomentsAlbum(mom.Title(), mom.Slug(), f.Serialize()); a != nil {
				if err := a.Create(); err != nil {
					log.Errorf("moments: %s", err.Error())
				} else {
					log.Infof("moments: added %s (%s)", txt.Quote(a.AlbumTitle), a.AlbumFilter)
				}
			} else {
				log.Errorf("moments: failed to create new moment %s (%s)", mom.Title(), f.Serialize())
			}
		}
	}

	if err := query.UpdateFolderDates(); err != nil {
		log.Errorf("moments: %s (update folder dates)", err.Error())
	}

	if err := query.UpdateAlbumDates(); err != nil {
		log.Errorf("moments: %s (update album dates)", err.Error())
	}

	if count, err := BackupAlbums(w.conf.AlbumsPath(), false); err != nil {
		log.Errorf("moments: %s (backup albums)", err.Error())
	} else if count > 0 {
		log.Debugf("moments: %d albums saved as yaml files", count)
	}

	return nil
}

// Cancel stops the current operation.
func (w *Moments) Cancel() {
	mutex.MainWorker.Cancel()
}
