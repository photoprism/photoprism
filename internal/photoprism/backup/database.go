package backup

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/log/status"
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

	// Storage-gate the on-disk path only: when toStdOut is true the dump
	// streams to stdout and does not consume the local storage volume, so an
	// operator can still offload a backup over ssh or similar when the disk
	// is full. Path validation and the InsufficientStorage check therefore
	// run inside this branch only.
	if !toStdOut {
		if backupPath == "" {
			backupPath = c.BackupDatabasePath()
		}

		// The same absolute paths are used for writing, rotation, and the cleanup of staged files.
		if backupPath, err = filepath.Abs(backupPath); err != nil {
			return err
		} else if fileName != "" {
			if fileName, err = filepath.Abs(fileName); err != nil {
				return err
			}
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

		// Refuse to write a new database dump if storage is over quota or critically low on free disk space.
		if c.InsufficientStorage() {
			return status.ErrInsufficientStorage
		}

		// Create backup path if not exists.
		if dir := filepath.Dir(fileName); dir != "." {
			if err = fs.MkdirAll(dir); err != nil {
				return err
			}
		}
	}

	var cmd *exec.Cmd

	// The password the command was built with, which its rendering must mask.
	var password string

	switch c.DatabaseDriver() {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		conn := newMariadbConn(c, c.MariadbDumpBin())
		logDatabaseSsl(conn, "backup")
		password, cmd = conn.Password, conn.Cmd()
	case dsn.DriverSQLite3:
		if !fs.FileExistsNotEmpty(c.DatabaseFile()) {
			return fmt.Errorf("sqlite database file %s not found", clean.LogQuote(c.DatabaseFile()))
		}

		cmd = exec.Command( // #nosec G204 sqlite dump uses configured binary and db path
			c.SqliteBin(),
			c.DatabaseFile(),
			".dump",
		)
	default:
		return fmt.Errorf("unsupported database type: %s", c.DatabaseDriver())
	}

	if toStdOut {
		log.Infof("backup: sending database backup to stdout")
		return runDump(cmd, os.Stdout, password)
	}

	log.Infof("backup: %s database backup file %s", backupAction, clean.Log(filepath.Base(fileName)))

	return writeDump(cmd, backupPath, fileName, password, force, retain)
}

// staleStageAge is the age after which a staged dump left by an interrupted run is removed.
const staleStageAge = 24 * time.Hour

// writeDump runs the dump command into a staged sibling of fileName and publishes it only once it
// completed, so a failed run leaves an existing file under that name untouched. Older dumps in
// backupPath are then rotated to retain.
func writeDump(cmd *exec.Cmd, backupPath, fileName, password string, force bool, retain int) (err error) {
	baseName := filepath.Base(fileName)

	if info, statErr := os.Lstat(fileName); statErr == nil {
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			return fmt.Errorf("%s is a symbolic link", clean.Log(baseName))
		case info.Mode().IsRegular():
		case isDumpName(baseName):
			return fmt.Errorf("%s is not a regular file", clean.Log(baseName))
		default:
			// A pipe or device named explicitly cannot be replaced by a rename, and holds no dump to keep.
			return writeDumpTo(cmd, fileName, password, os.Geteuid())
		}
	}

	removeStaleStages(filepath.Dir(fileName), staleStageAge)

	f, err := fs.OpenStageFileMode(fileName, fs.ModeBackupFile)

	if err != nil {
		return fmt.Errorf("failed to create %s (%s)", clean.Log(fileName), err)
	}

	stageName := f.Name()
	published := false

	defer func() {
		_ = f.Close()

		if !published {
			_ = os.Remove(stageName)
		}
	}()

	// Copying through a pipe returns write errors that some dump clients ignore.
	w := &dumpWriter{w: f}

	if err = runDump(cmd, w, password); w.err != nil {
		return w.err
	} else if err != nil {
		return err
	}

	if err = f.Sync(); err != nil {
		return err
	}

	info, err := f.Stat()

	if err != nil {
		return err
	} else if info.Size() == 0 {
		return fmt.Errorf("database dump for %s is empty", clean.Log(baseName))
	}

	if err = f.Close(); err != nil {
		return err
	}

	if err = fs.PublishFile(stageName, fileName, force); err != nil {
		return err
	}

	published = true

	return rotateDumps(backupPath, retain)
}

// writeDumpTo runs the dump command into an existing pipe or device owned by uid, without following
// a symlink.
func writeDumpTo(cmd *exec.Cmd, fileName, password string, uid int) error {
	// Checked before the open as well, since opening a pipe without a reader blocks.
	if info, err := os.Lstat(fileName); err == nil && !info.Mode().IsRegular() && !dumpTargetOwned(info, uid) {
		return fmt.Errorf("%s is not owned by the current user", clean.Log(filepath.Base(fileName)))
	}

	// #nosec G304 the name is an explicit destination checked by the caller
	f, err := os.OpenFile(fileName, os.O_WRONLY|fs.OpenNoFollow, 0)

	if err != nil {
		return fmt.Errorf("failed to open %s (%s)", clean.Log(fileName), err)
	}

	defer f.Close()

	// A regular file is written only through a staged sibling.
	if info, statErr := f.Stat(); statErr != nil {
		return statErr
	} else if info.Mode().IsRegular() {
		return fmt.Errorf("%s is a regular file", clean.Log(filepath.Base(fileName)))
	} else if !dumpTargetOwned(info, uid) {
		return fmt.Errorf("%s is not owned by the current user", clean.Log(filepath.Base(fileName)))
	}

	if err = runDump(cmd, f, password); err != nil {
		return err
	}

	return f.Close()
}

// dumpTargetOwned reports whether a pipe is owned by uid, or a device by uid or root.
func dumpTargetOwned(info os.FileInfo, uid int) bool {
	st, ok := info.Sys().(*syscall.Stat_t)

	if !ok {
		return false
	}

	owner := int(st.Uid)

	if info.Mode()&os.ModeNamedPipe != 0 {
		return owner == uid
	}

	return owner == uid || owner == 0
}

// isDumpName reports whether a base name matches the dated names of rotated database dumps.
func isDumpName(baseName string) bool {
	matched, _ := filepath.Match(SqlBackupFileNamePattern, baseName)
	return matched
}

// removeStaleStages removes staged dumps in dir that were last written more than maxAge ago.
// A younger stage may belong to a run in another process, so it is kept.
func removeStaleStages(dir string, maxAge time.Duration) {
	files, err := globIn(dir, "."+SqlBackupFileNamePattern+".*"+fs.ExtTmp+".sql")

	if err != nil {
		return
	}

	for _, name := range files {
		if info, statErr := os.Lstat(name); statErr != nil || !info.Mode().IsRegular() || time.Since(info.ModTime()) <= maxAge {
			continue
		} else if err = os.Remove(name); err != nil {
			log.Warnf("backup: failed to remove stale database backup file %s (%s)", clean.Log(filepath.Base(name)), err)
		} else {
			log.Infof("backup: removed stale database backup file %s", clean.Log(filepath.Base(name)))
		}
	}
}

// dumpWriter passes writes to w and records the first write error.
type dumpWriter struct {
	w   io.Writer
	err error
}

// Write writes p to the underlying writer and records the first error.
func (d *dumpWriter) Write(p []byte) (int, error) {
	n, err := d.w.Write(p)

	if err != nil && d.err == nil {
		d.err = err
	}

	return n, err
}

// runDump runs the dump command with its output sent to w, returning stderr as the error if it fails.
func runDump(cmd *exec.Cmd, w io.Writer, password string) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = w

	// Log the command for debugging in trace mode.
	log.Trace(clean.Cmd(cmd, password))

	if cmdErr := cmd.Run(); cmdErr != nil {
		if err := clientError(stderr.String(), password, "backup"); err != nil {
			return err
		}

		return cmdErr
	}

	clientDiagnostics(stderr.String(), password, "backup")

	return nil
}

// rotateDumps removes the oldest dumps in backupPath until only retain remain, if retain is set.
func rotateDumps(backupPath string, retain int) error {
	if backupPath == "" || retain <= 0 {
		return nil
	}

	files, err := globIn(backupPath, SqlBackupFileNamePattern)

	if err != nil {
		return err
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
		}

		log.Infof("backup: removed database backup file %s", clean.Log(filepath.Base(files[i])))
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

			files, globErr := globIn(backupPath, SqlBackupFileNamePattern)

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

	switch {
	case counts.Photos == 0:
		// No existing data to guard against.
	case !force:
		return fmt.Errorf("found an existing index with %d pictures, backup not restored", counts.Photos)
	default:
		log.Warnf("restore: existing index with %d pictures will be replaced", counts.Photos)
	}

	// The command is prepared before any table is dropped.
	cmd, password, err := restoreCmd(c)

	if err != nil {
		return err
	}

	if c.DatabaseDriver() == dsn.DriverSQLite3 {
		log.Infoln("restore: dropping existing sqlite database tables")
		entity.Entities.Drop(c.Db())
	}

	// Read from stdin or file.
	var f *os.File
	if fromStdIn {
		log.Infof("restore: restoring database backup from stdin")
		f = os.Stdin
		// #nosec G304 backup path validated by configuration
	} else if f, err = os.OpenFile(fileName, os.O_RDONLY, 0); err != nil {
		return fmt.Errorf("failed to open %s: %s", clean.Log(fileName), err)
	} else {
		log.Infof("restore: restoring database backup from %s", clean.Log(filepath.Base(fileName)))
		defer f.Close()
	}

	if err = runRestore(cmd, f, password); err != nil {
		log.Errorf("restore: failed to restore index database")
		return err
	}

	log.Infof("restore: index database successfully restored")

	return nil
}

// runRestore runs the restore command with its input read from r, returning stderr as the error if it fails.
// The input is copied through a pipe, so the client runs in batch mode even if r is a terminal, and the
// copy does not delay the result of a client that exits before reading all of it.
func runRestore(cmd *exec.Cmd, r io.Reader, password string) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	go func() {
		defer stdin.Close()

		if _, copyErr := io.Copy(stdin, r); copyErr != nil && !errors.Is(copyErr, syscall.EPIPE) {
			log.Errorf("restore: %s", clean.Error(copyErr))
		}
	}()

	// Log the command for debugging in trace mode.
	log.Trace(clean.Cmd(cmd, password))

	if cmdErr := cmd.Run(); cmdErr != nil {
		if err := clientError(stderr.String(), password, "restore"); err != nil {
			return err
		}

		return cmdErr
	}

	clientDiagnostics(stderr.String(), password, "restore")

	return nil
}
