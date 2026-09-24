package commands

import (
	"errors"
	"os"
	"syscall"

	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// StopCommand configures the command name, flags, and action.
var StopCommand = &cli.Command{
	Name:    "stop",
	Aliases: []string{"down"},
	Usage:   "Stops the Web server in daemon mode",
	Action:  stopAction,
}

// stopAction stops the daemon if it is running.
func stopAction(ctx *cli.Context) error {
	conf, err := InitCoreConfig(ctx, false)

	if err != nil {
		log.Debug(err)
	}

	log.Infof("looking for pid in %s", clean.Log(conf.PIDFilename()))

	dcxt := new(daemon.Context)
	dcxt.PidFileName = conf.PIDFilename()
	child, err := dcxt.Search()

	if err != nil {
		if daemonNotRunning(err, dcxt.PidFileName) {
			log.Info("daemon is not running")
			return nil
		}
		return err
	}

	if child == nil {
		log.Info("daemon is not running")
		return nil
	}

	err = child.Signal(syscall.SIGTERM)

	if err != nil {
		return err
	}

	st, err := child.Wait()

	if err != nil {
		log.Info("daemon exited successfully")
		return nil
	}

	log.Infof("daemon[%v] exited[%v]? successfully[%v]?\n", st.Pid(), st.Exited(), st.Success())

	return nil
}

// daemonNotRunning reports whether a failed pid file search means that no daemon is running.
// Search removes a pid file whose process has exited, which tells it apart from a read error.
func daemonNotRunning(err error, pidFile string) bool {
	return errors.Is(err, os.ErrNotExist) || !fs.FileExists(pidFile)
}
