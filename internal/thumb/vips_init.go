package thumb

import (
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/pkg/clean"
)

var (
	vipsStart   sync.Once
	vipsStarted = false
)

// VipsInit initializes libvips by checking its version and loading the ICC profiles once.
func VipsInit() {
	vipsStart.Do(vipsInit)
}

// VipsShutdown shuts down libvips and removes temporary files.
func VipsShutdown() {
	if vipsStarted {
		vipsStarted = false
		vips.Shutdown()
	}
}

// vipsInit calls vips.Startup() to initialize libvips.
func vipsInit() {
	if vipsStarted == true {
		log.Warnf("vips: already initialized - you may have found a bug")
		return
	}

	vipsStarted = true

	// Configure logging.
	vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
		switch level {
		case vips.LogLevelError, vips.LogLevelCritical:
			log.Errorf("vips: %s › %s", domain, clean.Log(msg))
		case vips.LogLevelWarning:
			log.Warnf("vips: %s › %s", domain, clean.Log(msg))
		case vips.LogLevelInfo, vips.LogLevelMessage:
			log.Infof("vips: %s › %s", domain, clean.Log(msg))
		default:
			log.Tracef("vips: %s › %s", domain, clean.Log(msg))
		}
	}, vipsLogLevel())

	// Start libvips.
	vips.Startup(vipsConfig())
}

// vipsConfig provides the config for initializing libvips.
func vipsConfig() *vips.Config {
	traceMode := log.GetLevel() == logrus.TraceLevel

	return &vips.Config{
		ConcurrencyLevel: ConcurrencyLevel,
		MaxCacheFiles:    MaxCacheFiles,
		MaxCacheMem:      MaxCacheMem,
		MaxCacheSize:     MaxCacheSize,
		ReportLeaks:      traceMode,
		CacheTrace:       false,
		CollectStats:     traceMode,
	}
}

// vipsLogLevel provides the libvips equivalent of the current log level.
func vipsLogLevel() vips.LogLevel {
	switch log.GetLevel() {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		return vips.LogLevelError
	case logrus.WarnLevel:
		return vips.LogLevelWarning
	case logrus.InfoLevel:
		return vips.LogLevelMessage
	default:
		return vips.LogLevelDebug
	}
}
