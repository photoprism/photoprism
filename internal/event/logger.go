package event

import (
	"context"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is a logrus compatible logger interface.
type Logger interface {
	WithField(key string, value any) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
	WithError(err error) *logrus.Entry
	WithContext(ctx context.Context) *logrus.Entry
	WithTime(t time.Time) *logrus.Entry
	Logf(level logrus.Level, format string, args ...any)
	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Printf(format string, args ...any)
	Warnf(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)
	Log(level logrus.Level, args ...any)
	LogFn(level logrus.Level, fn logrus.LogFunction)
	Trace(args ...any)
	Debug(args ...any)
	Info(args ...any)
	Print(args ...any)
	Warn(args ...any)
	Warning(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Panic(args ...any)
	TraceFn(fn logrus.LogFunction)
	DebugFn(fn logrus.LogFunction)
	InfoFn(fn logrus.LogFunction)
	PrintFn(fn logrus.LogFunction)
	WarnFn(fn logrus.LogFunction)
	WarningFn(fn logrus.LogFunction)
	ErrorFn(fn logrus.LogFunction)
	FatalFn(fn logrus.LogFunction)
	PanicFn(fn logrus.LogFunction)
	Logln(level logrus.Level, args ...any)
	Traceln(args ...any)
	Debugln(args ...any)
	Infoln(args ...any)
	Println(args ...any)
	Warnln(args ...any)
	Warningln(args ...any)
	Errorln(args ...any)
	Fatalln(args ...any)
	Panicln(args ...any)
	Exit(code int)
	SetNoLock()
	SetLevel(level logrus.Level)
	GetLevel() logrus.Level
	AddHook(hook logrus.Hook)
	IsLevelEnabled(level logrus.Level) bool
	SetFormatter(formatter logrus.Formatter)
	SetOutput(output io.Writer)
	SetReportCaller(reportCaller bool)
	ReplaceHooks(hooks logrus.LevelHooks) logrus.LevelHooks
	SetBufferPool(pool logrus.BufferPool)
}
