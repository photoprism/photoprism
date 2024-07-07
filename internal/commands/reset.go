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
	"github.com/photoprism/photoprism/internal/entity/migrate"
)

// ResetCommand configures the command name, flags, and action.
var ResetCommand = cli.Command{
	Name:  "reset",
	Usage: "Resets the index, clears the cache, and removes sidecar files",
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
			Usage: "assume \"yes\" and run non-interactively",
		},
	},
	Action: resetAction,
}

// resetAction resets the index and removes sidecar files after confirmation.
func resetAction(ctx *cli.Context) error {
	conf, err := InitConfig(ctx)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err != nil {
		return err
	}

	defer conf.Shutdown()

	if !ctx.Bool("yes") {
		log.Warnf("This will delete and recreate your index database after confirmation")

		if !ctx.Bool("index") {
			log.Warnf("You will be asked next if you also want to remove cache and sidecar files")
		}
	}

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("reset: enabled trace mode")
	}

	confirmed := ctx.Bool("yes")

	// Show prompt?
	if !confirmed {
		removeIndexPrompt := promptui.Prompt{
			Label:     "Delete and recreate index database?",
			IsConfirm: true,
		}

		if _, err := removeIndexPrompt.Run(); err == nil {
			confirmed = true
		} else {
			log.Infof("keeping index database")
		}
	}

	// Reset index?
	if confirmed {
		resetIndexDb(conf)
	}

	// Reset index only?
	if ctx.Bool("index") || ctx.Bool("yes") {
		return nil
	}

	// Clear cache.
	removeCachePrompt := promptui.Prompt{
		Label:     "Clear cache incl thumbnails?",
		IsConfirm: true,
	}

	if _, err := removeCachePrompt.Run(); err == nil {
		resetCache(conf)
	} else {
		log.Infof("keeping cache files")
	}

	// *.json sidecar files.
	removeSidecarJsonPrompt := promptui.Prompt{
		Label:     "Delete all *.json sidecar files?",
		IsConfirm: true,
	}

	if _, err := removeSidecarJsonPrompt.Run(); err == nil {
		resetSidecarJson(conf)
	} else {
		log.Infof("keeping *.json sidecar files")
	}

	// *.yml metadata files.
	removeSidecarYamlPrompt := promptui.Prompt{
		Label:     "Delete all *.yml metadata files?",
		IsConfirm: true,
	}

	if _, err := removeSidecarYamlPrompt.Run(); err == nil {
		resetSidecarYaml(conf)
	} else {
		log.Infof("keeping *.yml metadata files")
	}

	// *.yml album files.
	removeAlbumYamlPrompt := promptui.Prompt{
		Label:     "Delete all *.yml album files?",
		IsConfirm: true,
	}

	if _, err := removeAlbumYamlPrompt.Run(); err == nil {
		start := time.Now()

		matches, err := filepath.Glob(regexp.QuoteMeta(conf.BackupAlbumsPath()) + "/**/*.yml")

		if err != nil {
			return err
		}

		if len(matches) > 0 {
			log.Infof("%d *.yml album files will be removed", len(matches))

			for _, name := range matches {
				if err := os.Remove(name); err != nil {
					fmt.Print("E")
				} else {
					fmt.Print(".")
				}
			}

			fmt.Println("")

			log.Infof("removed all *.yml album files [%s]", time.Since(start))
		} else {
			log.Infof("found no *.yml album files")
		}
	} else {
		log.Infof("keeping *.yml album files")
	}

	return nil
}

// resetIndexDb resets the index database schema.
func resetIndexDb(c *config.Config) {
	start := time.Now()

	tables := entity.Entities

	log.Infoln("dropping existing tables")
	tables.Drop(c.Db())

	log.Infoln("restoring default schema")
	entity.InitDb(migrate.Opt(true, false, nil))

	// Reset admin account?
	if c.AdminPassword() == "" {
		log.Warnf("password required to reset admin account")
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword())
	}

	log.Infof("completed in %s", time.Since(start))
}

// resetCache removes all cache files and folders.
func resetCache(c *config.Config) {
	start := time.Now()

	matches, err := filepath.Glob(regexp.QuoteMeta(c.CachePath()) + "/**")

	if err != nil {
		log.Errorf("reset: %s (find cache files)", err)
		return
	}

	if len(matches) > 0 {
		log.Infof("clearing cache")

		for _, name := range matches {
			if err := os.RemoveAll(name); err != nil {
				fmt.Print("E")
			} else {
				fmt.Print(".")
			}
		}

		fmt.Println("")

		log.Infof("removed cache files [%s]", time.Since(start))
	} else {
		log.Infof("found no cache files")
	}
}

// resetSidecarJson removes generated *.json sidecar files.
func resetSidecarJson(c *config.Config) {
	start := time.Now()

	matches, err := filepath.Glob(regexp.QuoteMeta(c.SidecarPath()) + "/**/*.json")

	if err != nil {
		log.Errorf("reset: %s (find *.json sidecar files)", err)
		return
	}

	if len(matches) > 0 {
		log.Infof("removing %d *.json sidecar files", len(matches))

		for _, name := range matches {
			if err = os.Remove(name); err != nil {
				fmt.Print("E")
			} else {
				fmt.Print(".")
			}
		}

		fmt.Println("")

		log.Infof("removed *.json sidecar files [%s]", time.Since(start))
	} else {
		log.Infof("found no *.json sidecar files")
	}
}

// resetSidecarYaml removes generated *.yml files.
func resetSidecarYaml(c *config.Config) {
	start := time.Now()

	matches, err := filepath.Glob(regexp.QuoteMeta(c.SidecarPath()) + "/**/*.yml")

	if err != nil {
		log.Errorf("reset: %s (find *.yml metadata files)", err)
		return
	}

	if len(matches) > 0 {
		log.Infof("%d *.yml metadata files will be removed", len(matches))

		for _, name := range matches {
			if err := os.Remove(name); err != nil {
				fmt.Print("E")
			} else {
				fmt.Print(".")
			}
		}

		fmt.Println("")

		log.Infof("removed all *.yml metadata files [%s]", time.Since(start))
	} else {
		log.Infof("found no *.yml metadata files")
	}
}
