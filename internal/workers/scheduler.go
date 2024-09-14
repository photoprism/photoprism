package workers

import (
	"fmt"

	"github.com/go-co-op/gocron/v2"
)

var (
	Jobs      map[string]gocron.Job
	Scheduler gocron.Scheduler
)

func init() {
	Jobs = make(map[string]gocron.Job)
}

// NewJob adds a new background job to the scheduler, see https://pkg.go.dev/github.com/go-co-op/gocron/v2.
func NewJob(name, schedule string, function func(), parameters ...any) error {
	if schedule == "" {
		return nil
	}

	if name == "" {
		return fmt.Errorf("new job requires a name")
	}

	if function == nil {
		return fmt.Errorf("new job requires a task to run")
	}

	job, err := Scheduler.NewJob(
		gocron.CronJob(
			schedule,
			false,
		),
		gocron.NewTask(
			function,
			parameters...,
		),
		gocron.WithSingletonMode(gocron.LimitModeWait),
	)

	if err != nil {
		return err
	}

	Jobs[name] = job

	return nil
}
