package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/urfave/cli"
)

// ResetCommand resets the index and removes sidecar files after confirmation.
var ResetCommand = cli.Command{
	Name:   "reset",
	Usage:  "Resets the index and removes sidecar files after confirmation",
	Action: resetAction,
}

// resetAction resets the index and removes sidecar files after confirmation.
func resetAction(ctx *cli.Context) error {
	log.Warnf("'photoprism reset' resets the index and removes sidecar files after confirmation")

	removeIndexPrompt := promptui.Prompt{
		Label:     "Reset index database incl albums, labels, users and metadata?",
		IsConfirm: true,
	}

	conf := config.NewConfig(ctx)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	entity.SetDbProvider(conf)

	if _, err := removeIndexPrompt.Run(); err == nil {
		start := time.Now()

		tables := entity.Entities

		log.Infoln("dropping existing tables")
		tables.Drop()

		log.Infoln("restoring default schema")
		entity.MigrateDb(true)

		if conf.AdminPassword() != "" {
			log.Infoln("restoring initial admin password")
			entity.Admin.InitPassword(conf.AdminPassword())
		}

		log.Infof("database reset completed in %s", time.Since(start))
	} else {
		log.Infof("keeping index database")
	}

	removeSidecarJsonPrompt := promptui.Prompt{
		Label:     "Permanently delete all *.json photo sidecar files?",
		IsConfirm: true,
	}

	if _, err := removeSidecarJsonPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.SidecarPath()) + "/**/*.json")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("%d json photo sidecar files will be removed", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed json files in %s", time.Since(start))
		} else {
			log.Infof("no json files found")
		}
	} else {
		log.Infof("keeping json sidecar files")
	}

	removeSidecarYamlPrompt := promptui.Prompt{
		Label:     "Permanently delete all *.yml photo metadata backups?",
		IsConfirm: true,
	}

	if _, err := removeSidecarYamlPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.SidecarPath()) + "/**/*.yml")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("%d photo metadata backups will be removed", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed files in %s", time.Since(start))
		} else {
			log.Infof("no metadata backups found for removal")
		}
	} else {
		log.Infof("keeping backup files")
	}

	removeAlbumYamlPrompt := promptui.Prompt{
		Label:     "Permanently delete all *.yml album backups?",
		IsConfirm: true,
	}

	if _, err := removeAlbumYamlPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.AlbumsPath()) + "/**/*.yml")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("%d album backups will be removed", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed files in %s", time.Since(start))
		} else {
			log.Infof("no album backups found for removal")
		}
	} else {
		log.Infof("keeping backup files")
	}

	conf.Shutdown()

	return nil
}
