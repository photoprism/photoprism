package thumb

import (
	"strings"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/sirupsen/logrus"
)

var (
	vipsStarted = false
	vipsStart   = sync.Once{}
)

// VipsInit initializes libvips by checking its version and loading the ICC profiles once.
func VipsInit() {
	vipsStart.Do(vipsInit)
}

// VipsShutdown shuts down libvips and removes temporary files.
func VipsShutdown() {
	if vipsStarted {
		vipsStarted = false
		vipsStart = sync.Once{}
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
			log.Errorf("%s: %s", strings.TrimSpace(strings.ToLower(domain)), msg)
		case vips.LogLevelWarning:
			log.Warnf("%s: %s", strings.TrimSpace(strings.ToLower(domain)), msg)
		default:
			log.Tracef("%s: %s", strings.TrimSpace(strings.ToLower(domain)), msg)
		}
	}, vipsLogLevel())

	// Start libvips.
	vips.Startup(vipsConfig())
}

// vipsConfig provides the config for initializing libvips.
func vipsConfig() *vips.Config {
	return &vips.Config{
		MaxCacheMem:      MaxCacheMem,
		MaxCacheSize:     MaxCacheSize,
		MaxCacheFiles:    MaxCacheFiles,
		ConcurrencyLevel: NumWorkers,
		ReportLeaks:      false,
		CacheTrace:       false,
		CollectStats:     false,
	}
}

// vipsLogLevel provides the libvips equivalent of the current log level.
func vipsLogLevel() vips.LogLevel {
	switch log.GetLevel() {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		return vips.LogLevelError
	case logrus.TraceLevel:
		return vips.LogLevelDebug
	default:
		return vips.LogLevelWarning
	}
}
