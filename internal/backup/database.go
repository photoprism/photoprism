package backup

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Database creates a database backup dump with the specified file and path name.
func Database(backupPath, fileName string, toStdOut, force bool, retain int) (err error) {
	// Ensure that only one database backup/restore operation is running at a time.
	backupDatabaseMutex.Lock()
	defer backupDatabaseMutex.Unlock()

	// Backup action shown in logs.
	backupAction := "creating"

	// Get configuration.
	c := get.Config()

	if !toStdOut {
		if backupPath == "" {
			backupPath = c.BackupDatabasePath()
		}

		// Create the backup path if it does not already exist.
		if err = fs.MkdirAll(backupPath); err != nil {
			return err
		}

		// Check if the backup path is writable.
		if !fs.PathWritable(backupPath) {
			return fmt.Errorf("backup path is not writable")
		}

		if fileName == "" {
			backupFile := time.Now().UTC().Format("2006-01-02") + ".sql"
			fileName = filepath.Join(backupPath, backupFile)
		}

		log.Debugf("backup: database backups will be stored in %s", clean.Log(backupPath))

		if _, err = os.Stat(fileName); err == nil && !force {
			return fmt.Errorf("%s already exists", clean.Log(filepath.Base(fileName)))
		} else if err == nil {
			backupAction = "replacing"
		}

		// Create backup path if not exists.
		if dir := filepath.Dir(fileName); dir != "." {
			if err = fs.MkdirAll(dir); err != nil {
				return err
			}
		}
	}

	var cmd *exec.Cmd

	switch c.DatabaseDriver() {
	case config.MySQL, config.MariaDB:
		// Connect via Unix Domain Socket?
		if strings.HasPrefix(c.DatabaseServer(), "/") {
			cmd = exec.Command(
				c.MariadbDumpBin(),
				"--protocol", "socket",
				"-S", c.DatabaseServer(),
				"-u", c.DatabaseUser(),
				"-p"+c.DatabasePassword(),
				c.DatabaseName(),
			)
		} else {
			cmd = exec.Command(
				c.MariadbDumpBin(),
				"--protocol", "tcp",
				"-h", c.DatabaseHost(),
				"-P", c.DatabasePortString(),
				"-u", c.DatabaseUser(),
				"-p"+c.DatabasePassword(),
				c.DatabaseName(),
			)
		}
	case config.SQLite3:
		if !fs.FileExistsNotEmpty(c.DatabaseFile()) {
			return fmt.Errorf("sqlite database file %s not found", clean.LogQuote(c.DatabaseFile()))
		}

		cmd = exec.Command(
			c.SqliteBin(),
			c.DatabaseFile(),
			".dump",
		)
	default:
		return fmt.Errorf("unsupported database type: %s", c.DatabaseDriver())
	}

	// Write to stdout or file.
	var f *os.File
	if toStdOut {
		log.Infof("backup: sending database backup to stdout")
		f = os.Stdout
	} else if f, err = os.OpenFile(fileName, os.O_TRUNC|os.O_RDWR|os.O_CREATE, fs.ModeBackup); err != nil {
		return fmt.Errorf("failed to create %s (%s)", clean.Log(fileName), err)
	} else {
		log.Infof("backup: %s database backup file %s", backupAction, clean.Log(filepath.Base(fileName)))
		defer f.Close()
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = f

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run backup command.
	if cmdErr := cmd.Run(); cmdErr != nil {
		if errStr := strings.TrimSpace(stderr.String()); errStr != "" {
			return errors.New(errStr)
		}

		return cmdErr
	}

	// Delete old backups if the number of backup files to keep has been specified.
	if !toStdOut && backupPath != "" && retain > 0 {
		files, globErr := filepath.Glob(filepath.Join(regexp.QuoteMeta(backupPath), SqlBackupFileNamePattern))

		if globErr != nil {
			return globErr
		}

		if len(files) == 0 {
			return fmt.Errorf("found no database backup files in %s", backupPath)
		} else if len(files) <= retain {
			return nil
		}

		sort.Strings(files)

		log.Infof("backup: retaining %s", english.Plural(retain, "database backup", "database backups"))

		for i := 0; i < len(files)-retain; i++ {
			if err = os.Remove(files[i]); err != nil {
				return err
			} else {
				log.Infof("backup: removed database backup file %s", clean.Log(filepath.Base(files[i])))
			}
		}
	}

	return nil
}

// RestoreDatabase restores the database from a backup file with the specified path and name.
func RestoreDatabase(backupPath, fileName string, fromStdIn, force bool) (err error) {
	// Ensure that only one database backup/restore operation is running at a time.
	backupDatabaseMutex.Lock()
	defer backupDatabaseMutex.Unlock()

	c := get.Config()

	// If empty, use default backup file name.
	if !fromStdIn {
		if fileName == "" {
			if backupPath == "" {
				backupPath = c.BackupDatabasePath()
			}

			files, globErr := filepath.Glob(filepath.Join(regexp.QuoteMeta(backupPath), SqlBackupFileNamePattern))

			if globErr != nil {
				return globErr
			}

			if len(files) == 0 {
				return fmt.Errorf("failed to find a backup in %s, index cannot be restored", backupPath)
			}

			sort.Strings(files)

			fileName = files[len(files)-1]

			if !fs.FileExistsNotEmpty(fileName) {
				return fmt.Errorf("failed to open %s, index cannot be restored", filepath.Base(fileName))
			}
		} else if backupPath == "" {
			if absName, absErr := filepath.Abs(fileName); absErr == nil && fs.FileExists(absName) {
				fileName = absName
			} else if dir := filepath.Dir(fileName); dir != "" && dir != "." {
				return fmt.Errorf("failed to find %s, index cannot be restored", clean.Log(fileName))
			} else if absName = filepath.Join(c.BackupDatabasePath(), fileName); !fs.FileExists(absName) {
				return fmt.Errorf("failed to find %s in the %s backup path, index cannot be restored", clean.Log(fileName), clean.Log(filepath.Base(c.BackupDatabasePath())))
			} else {
				fileName = absName
			}
		} else if absName, absErr := filepath.Abs(filepath.Join(backupPath, fileName)); absErr == nil && fs.FileExists(absName) {
			fileName = absName
		} else {
			return fmt.Errorf("failed to find %s in %s, index cannot be restored", clean.Log(filepath.Base(fileName)), clean.Log(backupPath))
		}
	}

	counts := struct{ Photos int }{}

	c.Db().Unscoped().Table("photos").
		Select("COUNT(*) AS photos").
		Take(&counts)

	if counts.Photos == 0 {
		// Do nothing;
	} else if !force {
		return fmt.Errorf("found an existing index with %d pictures, backup not restored", counts.Photos)
	} else {
		log.Warnf("restore: existing index with %d pictures will be replaced", counts.Photos)
	}

	tables := entity.Entities

	var cmd *exec.Cmd

	switch c.DatabaseDriver() {
	case config.MySQL, config.MariaDB:
		cmd = exec.Command(
			c.MariadbBin(),
			"--protocol", "tcp",
			"-h", c.DatabaseHost(),
			"-P", c.DatabasePortString(),
			"-u", c.DatabaseUser(),
			"-p"+c.DatabasePassword(),
			"-f",
			c.DatabaseName(),
		)
	case config.SQLite3:
		log.Infoln("restore: dropping existing sqlite database tables")
		tables.Drop(c.Db())
		cmd = exec.Command(
			c.SqliteBin(),
			c.DatabaseFile(),
		)
	default:
		return fmt.Errorf("unsupported database type: %s", c.DatabaseDriver())
	}

	// Read from stdin or file.
	var f *os.File
	if fromStdIn {
		log.Infof("restore: restoring database backup from stdin")
		f = os.Stdin
	} else if f, err = os.OpenFile(fileName, os.O_RDONLY, 0); err != nil {
		return fmt.Errorf("failed to open %s: %s", clean.Log(fileName), err)
	} else {
		log.Infof("restore: restoring database backup from %s", clean.Log(filepath.Base(fileName)))
		defer f.Close()
	}

	var stderr bytes.Buffer
	var stdin io.WriteCloser
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout
	stdin, err = cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		if _, err = io.Copy(stdin, f); err != nil {
			log.Errorf(err.Error())
		}
	}()

	// Log exact command for debugging in trace mode.
	log.Trace(cmd.String())

	// Run restore command.
	if cmdErr := cmd.Run(); cmdErr != nil {
		log.Errorf("restore: failed to restore index database")

		if errStr := strings.TrimSpace(stderr.String()); errStr != "" {
			return errors.New(errStr)
		}

		return cmdErr
	} else {
		log.Infof("restore: index database successfully restored")
	}

	return nil
}
