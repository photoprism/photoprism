package backup

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

var (
	// clientProbeTimeout bounds the run that checks whether a client accepts a flag.
	clientProbeTimeout = 10 * time.Second
	// clientProbeWaitDelay bounds the wait for output after the probe was stopped, since a child
	// process may keep the output open.
	clientProbeWaitDelay = 2 * time.Second
)

// clientProbe is the cached result of running a client with a flag.
type clientProbe struct {
	accepted bool
	err      error
}

// clientProbes caches probe results, keyed by binary and arguments.
var clientProbes sync.Map

// clientAccepts reports whether bin accepts the arguments, running it once and caching the result.
// Clients report an unknown option either with an error exit or in their output, so both are checked.
// The error is set only if the client could not be run or did not finish in time.
func clientAccepts(bin string, args ...string) (bool, error) {
	key := bin + "\x00" + strings.Join(args, "\x00")

	if v, ok := clientProbes.Load(key); ok {
		p := v.(clientProbe)
		return p.accepted, p.err
	}

	ctx, cancel := context.WithTimeout(context.Background(), clientProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...) // #nosec G204 configured client binary
	cmd.WaitDelay = clientProbeWaitDelay

	out, runErr := cmd.CombinedOutput()

	var p clientProbe
	var exitErr *exec.ExitError

	switch {
	case ctx.Err() != nil:
		p.err = ctx.Err()
	case runErr == nil, errors.Is(runErr, exec.ErrWaitDelay):
		// A client that exited while a child kept its output open still ran.
		p.accepted = !strings.Contains(strings.ToLower(string(out)), "unknown option")
	case !errors.As(runErr, &exitErr):
		p.err = runErr
	}

	clientProbes.Store(key, p)

	return p.accepted, p.err
}

// mariadbRestoreArgs returns the flags the MariaDB client restores a dump with: binary mode, without
// loading local files, and in the client's sandbox mode where supported.
func mariadbRestoreArgs(bin string) []string {
	args := []string{"-f", "--local-infile=0", "--binary-mode"}

	if ok, err := clientAccepts(bin, "--no-defaults", "--sandbox", "--version"); ok {
		args = append(args, "--sandbox")
	} else if err != nil {
		log.Warnf("restore: failed to check %s for sandbox mode (%s)", filepath.Base(bin), clean.Error(err))
	} else {
		log.Warnf("restore: %s does not support sandbox mode", filepath.Base(bin))
	}

	return args
}

// sqliteRestoreCmd returns the command that restores a dump into dbFile, in the client's safe mode
// where supported.
func sqliteRestoreCmd(bin, dbFile string) *exec.Cmd {
	if ok, err := clientAccepts(bin, "-safe", "-version"); ok {
		return exec.Command(bin, "-safe", dbFile) // #nosec G204 configured binary and database path
	} else if err != nil {
		log.Warnf("restore: failed to check %s for safe mode (%s)", filepath.Base(bin), clean.Error(err))
	} else {
		log.Warnf("restore: %s does not support safe mode", filepath.Base(bin))
	}

	return exec.Command(bin, dbFile) // #nosec G204 configured binary and database path
}

// restoreCmd returns the command that restores a dump into the configured database, and the password
// its rendering must mask.
func restoreCmd(c *config.Config) (cmd *exec.Cmd, password string, err error) {
	switch c.DatabaseDriver() {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		if c.MariadbBin() == "" {
			return nil, "", errors.New("mariadb client not found")
		}

		conn := newMariadbConn(c, c.MariadbBin())
		logDatabaseSsl(conn, "restore")
		cmd, password = conn.Cmd(mariadbRestoreArgs(conn.Bin)...), conn.Password
	case dsn.DriverSQLite3:
		if c.SqliteBin() == "" {
			return nil, "", errors.New("sqlite3 client not found")
		}

		cmd = sqliteRestoreCmd(c.SqliteBin(), c.DatabaseFile())
	default:
		return nil, "", fmt.Errorf("unsupported database type: %s", c.DatabaseDriver())
	}

	// A client that was not found is reported before the restore changes anything.
	if cmd.Err != nil {
		return nil, "", cmd.Err
	}

	return cmd, password, nil
}
