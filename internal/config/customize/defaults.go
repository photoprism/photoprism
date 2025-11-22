package customize

import (
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/time/tz"
)

var (
	// DefaultTheme specifies the UI theme used when a user has not selected one.
	DefaultTheme = "default"
	// DefaultStartPage defines the default page shown after login.
	DefaultStartPage = "default"
	// DefaultMapsStyle is the map provider style used unless overridden by the user.
	DefaultMapsStyle = "default"
	// DefaultLanguage is the fallback locale for the UI.
	DefaultLanguage = i18n.Default.Locale()
	// DefaultTimeZone is the default timezone applied to user sessions.
	DefaultTimeZone = tz.Local
)
