package commands

import (
	"context"
	"errors"
	"fmt"
	gofs "io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// ResetCommand configures the command name, flags, and action.
var ResetCommand = &cli.Command{
	Name:  "reset",
	Usage: "Resets the index, clears the cache, and removes sidecar files",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "index",
			Aliases: []string{"i"},
			Usage:   "resets only the index database",
		},
		&cli.BoolFlag{
			Name:    "trace",
			Aliases: []string{"t"},
			Usage:   "shows trace logs for debugging",
		},
		&cli.BoolFlag{
			Name:    "yes",
			Aliases: []string{"y"},
			Usage:   "runs the command non-interactively",
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

	if !RunNonInteractively(ctx.Bool("yes")) {
		log.Warnf("This will delete and recreate your index database after confirmation")

		if !ctx.Bool("index") {
			log.Warnf("You will be asked next if you also want to remove cache and sidecar files")
		}
	}

	if ctx.Bool("trace") {
		log.SetLevel(logrus.TraceLevel)
		log.Infoln("reset: enabled trace mode")
	}

	confirmed := RunNonInteractively(ctx.Bool("yes"))

	// Show prompt?
	if !confirmed {
		removeIndexPrompt := promptui.Prompt{
			Label:     "Delete and recreate index database?",
			IsConfirm: true,
		}

		if _, err = removeIndexPrompt.Run(); err == nil {
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
	if RunNonInteractively(false) {
		log.Infof("keeping cache files")
	} else {
		removeCachePrompt := promptui.Prompt{Label: "Clear cache incl thumbnails?", IsConfirm: true}
		if _, err = removeCachePrompt.Run(); err == nil {
			resetCache(conf)
		} else {
			log.Infof("keeping cache files")
		}
	}

	// *.json sidecar files.
	if RunNonInteractively(false) {
		log.Infof("keeping *.json sidecar files")
	} else {
		removeSidecarJsonPrompt := promptui.Prompt{Label: "Delete all *.json sidecar files?", IsConfirm: true}
		if _, err = removeSidecarJsonPrompt.Run(); err == nil {
			resetSidecarJson(conf)
		} else {
			log.Infof("keeping *.json sidecar files")
		}
	}

	// *.yml metadata files.
	if RunNonInteractively(false) {
		log.Infof("keeping *.yml metadata files")
	} else {
		removeSidecarYamlPrompt := promptui.Prompt{Label: "Delete all *.yml metadata files?", IsConfirm: true}
		if _, err = removeSidecarYamlPrompt.Run(); err == nil {
			resetSidecarYaml(conf)
		} else {
			log.Infof("keeping *.yml metadata files")
		}
	}

	// *.yml album files.
	if !RunNonInteractively(false) {
		removeAlbumYamlPrompt := promptui.Prompt{Label: "Delete all *.yml album files?", IsConfirm: true}
		if _, err = removeAlbumYamlPrompt.Run(); err != nil {
			log.Infof("keeping *.yml album files")
			return nil
		}
	} else {
		log.Infof("keeping *.yml album files")
		return nil
	}

	resetAlbumYaml(conf)

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

	// A pinned face model only exists to keep new vectors comparable with the ones the library
	// already holds, and it now holds none, so the pin would outlive its reason.
	if err := c.ClearFaceModel(); err != nil {
		log.Warnf("reset: %s", err)
	}

	// Reset admin account?
	if c.AdminPassword() == "" {
		log.Warnf("password required to reset admin account")
	} else {
		entity.Admin.InitAccount(c.AdminUser(), c.AdminPassword(), c.AdminScope())
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

	entity.FlushCaches()
}

// resetSidecarJson removes generated *.json sidecar files.
func resetSidecarJson(c *config.Config) {
	resetSidecarFiles(c, fs.ExtJson, "*.json sidecar files")
}

// resetSidecarYaml removes generated *.yml files.
func resetSidecarYaml(c *config.Config) {
	resetSidecarFiles(c, fs.ExtYml, "*.yml metadata files")
}

// resetSidecarFiles removes the sidecar files with the given extension, at any depth.
// A relative sidecar path is skipped, since it would be resolved against the working directory,
// and so is one that equals or contains another library or config folder.
func resetSidecarFiles(c *config.Config, ext, kind string) {
	if !c.SidecarPathIsAbs() {
		log.Warnf("reset: removed 0 %s, because the sidecar path %s is relative", kind, clean.Log(c.SidecarPath()))
		return
	} else if overlap := resetSidecarOverlap(c); overlap != "" {
		log.Warnf("reset: removed 0 %s, because the sidecar path %s contains %s", kind, clean.Log(c.SidecarPath()), clean.Log(overlap))
		return
	}

	resetFiles(c.SidecarPath(), ext, kind)
}

// resetSidecarOverlap returns the library, backup, or config path that the sidecar path equals or
// contains, or "" if there is none. Symbolic links are resolved where the paths exist.
func resetSidecarOverlap(c *config.Config) string {
	sidecarPath := resolvedResetPath(c.SidecarPath())

	for _, dir := range []string{c.OriginalsPath(), c.ImportPath(), c.StoragePath(), c.ConfigPath(), c.BackupBasePath(), c.BackupAlbumsPath()} {
		if dir == "" {
			continue
		}

		rel, err := filepath.Rel(sidecarPath, resolvedResetPath(dir))

		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return dir
		}
	}

	return ""
}

// resolvedResetPath returns the cleaned path with symbolic links resolved, or just cleaned if it
// cannot be resolved.
func resolvedResetPath(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}

	return filepath.Clean(dir)
}

// resetAlbumYaml removes the *.yml album backup files.
func resetAlbumYaml(c *config.Config) {
	resetFiles(c.BackupAlbumsPath(), fs.ExtYml, "*.yml album files")
}

// resetFiles removes the files with the given extension below dir and logs how many were removed.
func resetFiles(dir, ext, kind string) {
	start := time.Now()

	log.Infof("removing %s in %s", kind, clean.Log(dir))

	removed, failed, err := removeFilesWithExt(dir, ext)

	if removed > 0 || failed > 0 {
		fmt.Println("")
	}

	if err != nil {
		log.Errorf("reset: %s (remove %s)", clean.Error(err), kind)
	}

	if failed > 0 {
		log.Errorf("reset: failed to remove %d %s", failed, kind)
	}

	if removed > 0 {
		log.Infof("removed %d %s [%s]", removed, kind, time.Since(start))
	} else if failed == 0 && err == nil {
		log.Infof("found no %s", kind)
	}
}

// removeFilesWithExt removes the regular files below dir whose extension is ext, at any depth, and
// prints one character per file as progress. Symbolic links are skipped, the os.Root keeps removal
// inside dir, and folders that cannot be read are skipped and returned as errors.
func removeFilesWithExt(dir, ext string) (removed, failed int, err error) {
	root, err := os.OpenRoot(dir)

	if errors.Is(err, gofs.ErrNotExist) {
		return 0, 0, nil
	} else if err != nil {
		return 0, 0, err
	}

	defer root.Close()

	var skipped []error

	err = gofs.WalkDir(root.FS(), ".", func(name string, d gofs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if name == "." {
				return walkErr
			}

			skipped = append(skipped, walkErr)
			return nil
		}

		if !d.Type().IsRegular() || path.Ext(name) != ext {
			return nil
		}

		if removeErr := root.Remove(name); removeErr != nil {
			failed++
			fmt.Print("E")
		} else {
			removed++
			fmt.Print(".")
		}

		return nil
	})

	return removed, failed, errors.Join(append(skipped, err)...)
}
