package dummy

//revive:disable:exported

import (
	"context"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger represents a dummy logger.
type Logger struct {
	Out          io.Writer
	Hooks        logrus.LevelHooks
	Formatter    logrus.Formatter
	ReportCaller bool
	Level        logrus.Level
	ExitFunc     exitFunc
	BufferPool   logrus.BufferPool
}

// Logger represents am exit callback function.
type exitFunc func(int)

// NewLogger creates a new dummy logger instance.
func NewLogger() *Logger {
	return &Logger{
		Out: os.Stderr,
		Formatter: &logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		},
		Hooks:    make(logrus.LevelHooks),
		Level:    logrus.PanicLevel,
		ExitFunc: os.Exit,
	}
}

// WithField allocates a new entry and adds a field to it.
func (logger *Logger) WithField(key string, value any) *logrus.Entry {
	return &logrus.Entry{Data: logrus.Fields{key: value}}
}

// WithFields adds a struct of fields to the log entry. All it does is call `WithField` for
// each `Field`.
func (logger *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return &logrus.Entry{Data: fields}
}

// WithError adds an error as single field to the log entry.  All it does is call
// `WithError` for the given `error`.
func (logger *Logger) WithError(err error) *logrus.Entry {
	if err == nil {
		return &logrus.Entry{}
	}
	return &logrus.Entry{Message: err.Error()}
}

// WithContext adds a context to the log entry.
func (logger *Logger) WithContext(ctx context.Context) *logrus.Entry {
	return &logrus.Entry{Context: ctx}
}

// WithTime overrides the time of the log entry.
func (logger *Logger) WithTime(t time.Time) *logrus.Entry {
	return &logrus.Entry{Time: t}
}

func (logger *Logger) Logf(level logrus.Level, format string, args ...any) {
}

func (logger *Logger) Tracef(format string, args ...any) {
}

func (logger *Logger) Debugf(format string, args ...any) {
}

func (logger *Logger) Infof(format string, args ...any) {
}

func (logger *Logger) Printf(format string, args ...any) {
}

func (logger *Logger) Warnf(format string, args ...any) {
}

func (logger *Logger) Warningf(format string, args ...any) {
}

func (logger *Logger) Errorf(format string, args ...any) {
}

func (logger *Logger) Fatalf(format string, args ...any) {
}

func (logger *Logger) Panicf(format string, args ...any) {
}

// Log will log a message at the level given as parameter.
func (logger *Logger) Log(level logrus.Level, args ...any) {
}

func (logger *Logger) LogFn(level logrus.Level, fn logrus.LogFunction) {
}

func (logger *Logger) Trace(args ...any) {
}

func (logger *Logger) Debug(args ...any) {
}

func (logger *Logger) Info(args ...any) {
}

func (logger *Logger) Print(args ...any) {
}

func (logger *Logger) Warn(args ...any) {
}

func (logger *Logger) Warning(args ...any) {
}

func (logger *Logger) Error(args ...any) {
}

func (logger *Logger) Fatal(args ...any) {
}

func (logger *Logger) Panic(args ...any) {
}

func (logger *Logger) TraceFn(fn logrus.LogFunction) {
}

func (logger *Logger) DebugFn(fn logrus.LogFunction) {
}

func (logger *Logger) InfoFn(fn logrus.LogFunction) {
}

func (logger *Logger) PrintFn(fn logrus.LogFunction) {
}

func (logger *Logger) WarnFn(fn logrus.LogFunction) {
}

func (logger *Logger) WarningFn(fn logrus.LogFunction) {
}

func (logger *Logger) ErrorFn(fn logrus.LogFunction) {
}

func (logger *Logger) FatalFn(fn logrus.LogFunction) {
}

func (logger *Logger) PanicFn(fn logrus.LogFunction) {
}

func (logger *Logger) Logln(level logrus.Level, args ...any) {
}

func (logger *Logger) Traceln(args ...any) {
}

func (logger *Logger) Debugln(args ...any) {
}

func (logger *Logger) Infoln(args ...any) {
	logger.Logln(logrus.InfoLevel, args...)
}

func (logger *Logger) Println(args ...any) {
}

func (logger *Logger) Warnln(args ...any) {
	logger.Logln(logrus.WarnLevel, args...)
}

func (logger *Logger) Warningln(args ...any) {
	logger.Warnln(args...)
}

func (logger *Logger) Errorln(args ...any) {
	logger.Logln(logrus.ErrorLevel, args...)
}

func (logger *Logger) Fatalln(args ...any) {
	logger.Logln(logrus.FatalLevel, args...)
}

func (logger *Logger) Panicln(args ...any) {
	logger.Logln(logrus.PanicLevel, args...)
}

func (logger *Logger) Exit(code int) {
}

func (logger *Logger) SetNoLock() {
}

// SetLevel sets the logger level.
func (logger *Logger) SetLevel(level logrus.Level) {
	atomic.StoreUint32((*uint32)(&logger.Level), uint32(level))
}

// GetLevel returns the logger level.
func (logger *Logger) GetLevel() logrus.Level {
	return logger.Level
}

// AddHook adds a hook to the logger hooks.
func (logger *Logger) AddHook(hook logrus.Hook) {
	logger.Hooks.Add(hook)
}

// IsLevelEnabled checks if the log level of the logger is greater than the level param
func (logger *Logger) IsLevelEnabled(level logrus.Level) bool {
	return logger.Level >= level
}

// SetFormatter sets the logger formatter.
func (logger *Logger) SetFormatter(formatter logrus.Formatter) {
	logger.Formatter = formatter
}

// SetOutput sets the logger output.
func (logger *Logger) SetOutput(output io.Writer) {
	logger.Out = output
}

func (logger *Logger) SetReportCaller(reportCaller bool) {
	logger.ReportCaller = reportCaller
}

// ReplaceHooks replaces the logger hooks and returns the old ones
func (logger *Logger) ReplaceHooks(hooks logrus.LevelHooks) logrus.LevelHooks {
	logger.Hooks = hooks
	return logger.Hooks
}

// SetBufferPool sets the logger buffer pool.
func (logger *Logger) SetBufferPool(pool logrus.BufferPool) {
	logger.BufferPool = pool
}
