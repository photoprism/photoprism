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

// ResetCommand resets the index and optionally removes YAML sidecar backup files.
var ResetCommand = cli.Command{
	Name:   "reset",
	Usage:  "Resets the index and optionally removes YAML sidecar backup files",
	Action: resetAction,
}

// resetAction removes the index and sidecar files after asking for confirmation.
func resetAction(ctx *cli.Context) error {
	log.Warnf("'photoprism reset' removes ALL data incl albums from the existing database")

	removeIndex := promptui.Prompt{
		Label:     "Reset index database?",
		IsConfirm: true,
	}

	_, err := removeIndex.Run()

	if err != nil {
		return fmt.Errorf("abort")
	}

	start := time.Now()

	conf := config.NewConfig(ctx)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	entity.SetDbProvider(conf)

	tables := entity.Entities

	log.Infoln("dropping existing tables")
	tables.Drop()

	log.Infoln("restoring default schema")
	entity.MigrateDb()

	if conf.AdminPassword() != "" {
		log.Infoln("restoring initial admin password")
		entity.Admin.InitPassword(conf.AdminPassword())
	}

	log.Infof("database reset completed in %s", time.Since(start))

	removeSidecarYaml := promptui.Prompt{
		Label:     "Permanently delete all *.yml photo metadata backups?",
		IsConfirm: true,
	}

	if _, err := removeSidecarYaml.Run(); err == nil {
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
			log.Infof("no backup files found")
		}
	} else {
		log.Infof("keeping backup files")
	}

	removeAlbumYaml := promptui.Prompt{
		Label:     "Permanently delete all *.yml album backups?",
		IsConfirm: true,
	}

	if _, err := removeAlbumYaml.Run(); err == nil {
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
			log.Infof("no backup files found")
		}
	} else {
		log.Infof("keeping backup files")
	}

	conf.Shutdown()

	return nil
}
