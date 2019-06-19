package commands

import (
	"syscall"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/prometheus/common/log"
	daemon "github.com/sevlyar/go-daemon"
	"github.com/urfave/cli"
)

// StopCommand stops the daemon if any.
var StopCommand = cli.Command{
	Name:   "stop",
	Usage:  "Stops daemon",
	Action: stopAction,
}

func stopAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	log.Infof("Looking for PID from file: %v\n", conf.DaemonPIDPath())
	dcxt := new(daemon.Context)
	dcxt.PidFileName = conf.DaemonPIDPath()
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
		log.Info("Daemon exited successfully")
		return nil
	}

	log.Infof("Daemon[%v] exited[%v]? successfully[%v]?\n", st.Pid(), st.Exited(), st.Success())
	return nil
}
