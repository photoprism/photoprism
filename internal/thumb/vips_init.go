package thumb

import (
	"strings"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/event"
)

var (
	vipsStarted = false
	vipsStopped = false
	vipsStart   = sync.Once{}
)

// VipsInit initializes libvips by checking its version and loading the ICC profiles once.
func VipsInit() {
	if vipsStopped {
		log.Debugf("vips: restart requested after shutdown, ignoring")
		return
	}

	vipsStart.Do(vipsInit)
}

// VipsShutdown shuts down libvips once as terminal process cleanup.
func VipsShutdown() {
	if !vipsStarted || vipsStopped {
		return
	}

	vipsStopped = true
	vips.Shutdown()
}

// vipsInit calls vips.Startup() to initialize libvips.
func vipsInit() {
	if vipsStopped {
		log.Debugf("vips: restart requested after shutdown, ignoring")
		return
	}

	if vipsStarted {
		log.Debugf("vips: already initialized")
		return
	}

	vipsStarted = true

	// Configure logging.
	vips.LoggingSettings(vipsLog, vipsLogLevel())

	// Start libvips.
	if err := vips.Startup(vipsConfig()); err != nil {
		vipsStarted = false
		log.Errorf("vips: %s", err)
	}
}

// vipsLog writes a libvips message to the log. Errors go to the system log, as libvips names the
// cache files it reads and writes.
func vipsLog(domain string, level vips.LogLevel, msg string) {
	domain = strings.TrimSpace(strings.ToLower(domain))

	switch level {
	case vips.LogLevelError, vips.LogLevelCritical:
		event.SystemError([]string{"vips", "%s: %s"}, domain, msg)
	case vips.LogLevelWarning:
		log.Debugf("%s: %s", domain, msg)
	default:
		log.Tracef("%s: %s", domain, msg)
	}
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
	case logrus.DebugLevel:
		return vips.LogLevelWarning
	case logrus.TraceLevel:
		return vips.LogLevelDebug
	default:
		return vips.LogLevelError
	}
}
