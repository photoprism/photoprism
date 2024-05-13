package photoprism

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

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const SqlBackupFileNamePattern = "[2-9][0-9][0-9][0-9]-[0-1][0-9]-[0-3][0-9].sql"

// RestoreIndex restores the index from an SQL backup dump with the specified file and path name.
func RestoreIndex(backupPath, fileName string, fromStdIn, force bool) (err error) {
	// Make sure only one backup/restore operation is running at a time.
	backupIndexMutex.Lock()
	defer backupIndexMutex.Unlock()

	c := Config()

	// If empty, use default backup file name.
	if !fromStdIn && fileName == "" {
		if backupPath == "" {
			backupPath = c.BackupIndexPath()
		}

		files, globErr := filepath.Glob(filepath.Join(regexp.QuoteMeta(backupPath), SqlBackupFileNamePattern))

		if globErr != nil {
			return globErr
		}

		if len(files) == 0 {
			return fmt.Errorf("found no backups files in %s", backupPath)
		}

		sort.Strings(files)

		fileName = files[len(files)-1]

		if !fs.FileExistsNotEmpty(fileName) {
			return fmt.Errorf("no backup found in %s", filepath.Base(fileName))
		}
	}

	counts := struct{ Photos int }{}

	c.Db().Unscoped().Table("photos").
		Select("COUNT(*) AS photos").
		Take(&counts)

	if counts.Photos == 0 {
		// Do nothing;
	} else if !force {
		return fmt.Errorf("found existing index with %d pictures, use the force option to replace it", counts.Photos)
	} else {
		log.Warnf("replacing the existing index with %d pictures", counts.Photos)
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
		log.Infoln("dropping existing tables")
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
		log.Infof("restoring index from stdin")
		f = os.Stdin
	} else if f, err = os.OpenFile(fileName, os.O_RDONLY, 0); err != nil {
		return fmt.Errorf("failed to open %s: %s", clean.Log(fileName), err)
	} else {
		log.Infof("restoring index from %s", clean.Log(fileName))
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
		log.Errorf("failed to restore index")

		if errStr := strings.TrimSpace(stderr.String()); errStr != "" {
			return errors.New(errStr)
		}

		return cmdErr
	}

	return nil
}
