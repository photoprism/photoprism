package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// ResetCommand resets the index and removes sidecar files after confirmation.
var ResetCommand = cli.Command{
	Name:  "reset",
	Usage: "Resets the index and removes generated sidecar files",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "index, i",
			Usage: "reset index database only",
		},
		cli.BoolFlag{
			Name:  "trace, t",
			Usage: "show trace logs for debugging",
		},
		cli.BoolFlag{
			Name:  "yes, y",
			Usage: "assume \"yes\" as answer to all prompts and run non-interactively",
		},
	},
	Action: resetAction,
}

// resetAction resets the index and removes sidecar files after confirmation.
func resetAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := conf.Init(); err != nil {
		return err
	}

	defer conf.Shutdown()

	entity.SetDbProvider(conf)

	if !ctx.Bool("yes") {
		log.Warnf("This will delete and recreate your index database after confirmation")

		if !ctx.Bool("index") {
			log.Warnf("You will be asked next if you also want to remove all JSON and YAML backup files")
		}
	}

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("reset: enabled trace mode")
	}

	resetIndex := ctx.Bool("yes")

	// Show prompt?
	if !resetIndex {
		removeIndexPrompt := promptui.Prompt{
			Label:     "Delete and recreate index database?",
			IsConfirm: true,
		}

		if _, err := removeIndexPrompt.Run(); err == nil {
			resetIndex = true
		} else {
			log.Infof("keeping index database")
		}
	}

	// Reset index?
	if resetIndex {
		resetIndexDb(conf)
	}

	// Reset index only?
	if ctx.Bool("index") {
		return nil
	}

	removeSidecarJsonPrompt := promptui.Prompt{
		Label:     "Delete all JSON metadata sidecar files?",
		IsConfirm: true,
	}

	if _, err := removeSidecarJsonPrompt.Run(); err == nil {
		resetSidecarJson(conf)
	} else {
		log.Infof("keeping JSON metadata sidecar files")
	}

	removeSidecarYamlPrompt := promptui.Prompt{
		Label:     "Delete all YAML metadata backup files?",
		IsConfirm: true,
	}

	if _, err := removeSidecarYamlPrompt.Run(); err == nil {
		resetSidecarYaml(conf)
	} else {
		log.Infof("keeping YAML metadata backups")
	}

	removeAlbumYamlPrompt := promptui.Prompt{
		Label:     "Delete all YAML album backup files?",
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

	return nil
}

// resetIndexDb resets the index database schema.
func resetIndexDb(conf *config.Config) {
	start := time.Now()

	tables := entity.Entities

	log.Infoln("dropping existing tables")
	tables.Drop(conf.Db())

	log.Infoln("restoring default schema")
	entity.MigrateDb(true, false)

	if conf.AdminPassword() != "" {
		log.Infoln("restoring initial admin password")
		entity.Admin.InitPassword(conf.AdminPassword())
	}

	log.Infof("database reset completed in %s", time.Since(start))
}

// resetSidecarJson removes generated JSON sidecar files.
func resetSidecarJson(conf *config.Config) {
	start := time.Now()

	matches, err := filepath.Glob(regexp.QuoteMeta(conf.SidecarPath()) + "/**/*.json")

	if err != nil {
		log.Errorf("reset: %s (find json sidecar files)", err)
		return
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
}

// resetSidecarYaml removes generated YAML sidecar files.
func resetSidecarYaml(conf *config.Config) {
	start := time.Now()

	matches, err := filepath.Glob(regexp.QuoteMeta(conf.SidecarPath()) + "/**/*.yml")

	if err != nil {
		log.Errorf("reset: %s (find yaml sidecar files)", err)
		return
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
}
