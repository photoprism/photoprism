package migrate

// Migration stages: StagePre runs before GORM AutoMigrate, StageMain after it, and StagePost last.
const (
	StagePre  = "pre"
	StageMain = "main"
	StagePost = "post"
)
