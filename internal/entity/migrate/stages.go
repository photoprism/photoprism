package migrate

// Migration stages: StagePre runs before GORM AutoMigrate, StageMain after it.
const (
	StagePre  = "pre"
	StageMain = "main"
	StagePost = "post"
)
