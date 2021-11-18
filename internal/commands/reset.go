package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// ResetCommand resets the index and removes sidecar files after confirmation.
var ResetCommand = cli.Command{
	Name:   "reset",
	Usage:  "Resets the index and removes generated sidecar files",
	Action: resetAction,
}

// resetAction resets the index and removes sidecar files after confirmation.
func resetAction(ctx *cli.Context) error {
	log.Warnf("YOU ARE ABOUT TO RESET THE INDEX AND REMOVE ALL JSON / YAML SIDECAR FILES")

	removeIndexPrompt := promptui.Prompt{
		Label:     "Reset index database including albums and metadata?",
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
		Label:     "Permanently delete existing JSON metadata sidecar files?",
		IsConfirm: true,
	}

	if _, err := removeSidecarJsonPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.SidecarPath()) + "/**/*.json")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("removing %d JSON metadata sidecar files", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed JSON metadata sidecar files [%s]", time.Since(start))
		} else {
			log.Infof("found no JSON metadata sidecar files")
		}
	} else {
		log.Infof("keeping JSON metadata sidecar files")
	}

	removeSidecarYamlPrompt := promptui.Prompt{
		Label:     "Permanently delete existing YAML metadata backups?",
		IsConfirm: true,
	}

	if _, err := removeSidecarYamlPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.SidecarPath()) + "/**/*.yml")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("%d YAML metadata backups will be removed", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed all YAML metadata backups [%s]", time.Since(start))
		} else {
			log.Infof("found no YAML metadata backups")
		}
	} else {
		log.Infof("keeping YAML metadata backups")
	}

	removeAlbumYamlPrompt := promptui.Prompt{
		Label:     "Permanently delete existing YAML album backups?",
		IsConfirm: true,
	}

	if _, err := removeAlbumYamlPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.AlbumsPath()) + "/**/*.yml")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("%d YAML album backups will be removed", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed all YAML album backups [%s]", time.Since(start))
		} else {
			log.Infof("found no YAML album backup files")
		}
	} else {
		log.Infof("keeping YAML album backup files")
	}

	conf.Shutdown()

	return nil
}
