package event

import (
	"context"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is a logrus compatible logger interface.
type Logger interface {
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
	WithError(err error) *logrus.Entry
	WithContext(ctx context.Context) *logrus.Entry
	WithTime(t time.Time) *logrus.Entry
	Logf(level logrus.Level, format string, args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Log(level logrus.Level, args ...interface{})
	LogFn(level logrus.Level, fn logrus.LogFunction)
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	TraceFn(fn logrus.LogFunction)
	DebugFn(fn logrus.LogFunction)
	InfoFn(fn logrus.LogFunction)
	PrintFn(fn logrus.LogFunction)
	WarnFn(fn logrus.LogFunction)
	WarningFn(fn logrus.LogFunction)
	ErrorFn(fn logrus.LogFunction)
	FatalFn(fn logrus.LogFunction)
	PanicFn(fn logrus.LogFunction)
	Logln(level logrus.Level, args ...interface{})
	Traceln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
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
