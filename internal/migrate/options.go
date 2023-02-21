package migrate

// Options represents migration selection options.
type Options struct {
	AutoMigrate    bool
	RunStage       string
	RunFailed      bool
	Migrations     []string
	DropDeprecated bool
}

// Opt returns migration options based on the specified parameters.
func Opt(runAll, runFailed bool, ids []string) Options {
	runAll = len(ids) == 0 && (runAll || runFailed)
	return Options{
		AutoMigrate:    runAll,
		RunStage:       StageMain,
		RunFailed:      runFailed,
		Migrations:     ids,
		DropDeprecated: runAll,
	}
}

// Stage returns options for the specified migration stage.
func (opt Options) Stage(name string) Options {
	if name == "" {
		return opt
	}

	return Options{
		AutoMigrate:    false,
		RunStage:       name,
		RunFailed:      opt.RunFailed,
		Migrations:     opt.Migrations,
		DropDeprecated: opt.DropDeprecated,
	}
}

// Pre returns options for the pre-migration stage.
func (opt Options) Pre() Options {
	return opt.Stage(StagePre)
}

// StageName returns the stage name.
func (opt Options) StageName() string {
	if opt.RunStage == "" {
		return StageMain
	} else {
		return opt.RunStage
	}
}
