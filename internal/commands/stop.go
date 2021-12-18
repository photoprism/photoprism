package commands

import (
	"syscall"

	"github.com/sevlyar/go-daemon"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// StopCommand registers the stop cli command.
var StopCommand = cli.Command{
	Name:    "stop",
	Aliases: []string{"down"},
	Usage:   "Stops the web server in daemon mode",
	Action:  stopAction,
}

// stopAction stops the daemon if it is running.
func stopAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	log.Infof("looking for pid in %s", sanitize.Log(conf.PIDFilename()))

	dcxt := new(daemon.Context)
	dcxt.PidFileName = conf.PIDFilename()
	child, err := dcxt.Search()

	if err != nil {
		log.Fatal(err)
	}

	err = child.Signal(syscall.SIGTERM)

	if err != nil {
		log.Fatal(err)
	}

	st, err := child.Wait()

	if err != nil {
		log.Info("daemon exited successfully")
		return nil
	}

	log.Infof("daemon[%v] exited[%v]? successfully[%v]?\n", st.Pid(), st.Exited(), st.Success())

	return nil
}
