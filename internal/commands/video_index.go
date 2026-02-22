package commands

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

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

	index := get.Index()
	result := photoprism.IndexRelated(related, index, photoprism.IndexOptionsSingle(conf))
	if result.Err != nil {
		return result.Err
	}

	return nil
}
