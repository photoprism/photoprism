package commands

import (
	"fmt"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
)

// videoFacesLocked returns the face migration holding the lock, or ""; tests replace it.
var videoFacesLocked = func(conf *config.Config) string {
	return conf.FacesLocked()
}

// videoMigrationRefused returns the exit error for a video command that would change a file while a
// face migration holds the lock, or nil if none does. Once done files of the batch were processed, it
// says to index them rather than to repeat the command, which would change them a second time.
func videoMigrationRefused(conf *config.Config, action string, done int) error {
	if conf == nil {
		return nil
	}

	held := videoFacesLocked(conf)

	if held == "" {
		return nil
	} else if done < 1 {
		return faceMigrationHeldExit(action, held)
	}

	return cli.Exit(fmt.Sprintf("%s: stopped for the %s after %s, run \"photoprism index\" once it has completed "+
		"and run the command again only for the files it has not processed", action, held, english.Plural(done, "file", "files")), 1)
}

// videoReindexRelated reindexes the related file group for the given main file.
func videoReindexRelated(conf *config.Config, fileName string) error {
	if fileName == "" {
		return fmt.Errorf("index: missing filename")
	}

	mediaFile, err := photoprism.NewMediaFile(fileName)
	if err != nil {
		return err
	}

	related, err := mediaFile.RelatedFiles(conf.Settings().Stack.Name)
	if err != nil {
		return err
	}

	// Related files whose names are not portable are not indexed with the file.
	if related = photoprism.PortableSiblings(related, mediaFile, func(f *photoprism.MediaFile) {
		log.Debugf("index: skipped %s because its name is not supported", clean.Log(f.BaseName()))
	}); related.Main == nil {
		return nil
	}

	// A migration that started since the command checked the lock leaves the file to the next index.
	if held := conf.FacesLocked(); held != "" {
		return cli.Exit(fmt.Sprintf("index: waiting for the %s to complete, run \"photoprism index\" once it has", held), 1)
	}

	index := get.Index()
	result := photoprism.IndexRelated(related, index, photoprism.IndexOptionsSingle(conf))
	if result.Err != nil {
		return result.Err
	}

	return nil
}
