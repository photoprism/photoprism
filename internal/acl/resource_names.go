package acl

// Resources that Roles can be granted Permission.
const (
	ResourceFiles     Resource = "files"
	ResourceFolders   Resource = "folders"
	ResourceShares    Resource = "shares"
	ResourcePhotos    Resource = "photos"
	ResourceVideos    Resource = "videos"
	ResourceFavorites Resource = "favorites"
	ResourceAlbums    Resource = "albums"
	ResourceMoments   Resource = "moments"
	ResourceCalendar  Resource = "calendar"
	ResourcePeople    Resource = "people"
	ResourcePlaces    Resource = "places"
	ResourceLabels    Resource = "labels"
	ResourceConfig    Resource = "config"
	ResourceSettings  Resource = "settings"
	ResourcePasscode  Resource = "passcode"
	ResourcePassword  Resource = "password"
	ResourceServices  Resource = "services"
	ResourceUsers     Resource = "users"
	ResourceSessions  Resource = "sessions"
	ResourceLogs      Resource = "logs"
	ResourceWebDAV    Resource = "webdav"
	ResourceMetrics   Resource = "metrics"
	ResourceFeedback  Resource = "feedback"
	ResourceDefault   Resource = "default"
)

// ResourceNames contains a list of all specified resources.
var ResourceNames = []Resource{
	ResourceFiles,
	ResourceFolders,
	ResourceShares,
	ResourcePhotos,
	ResourceVideos,
	ResourceFavorites,
	ResourceAlbums,
	ResourceMoments,
	ResourceCalendar,
	ResourcePeople,
	ResourcePlaces,
	ResourceLabels,
	ResourceConfig,
	ResourceSettings,
	ResourcePasscode,
	ResourcePassword,
	ResourceServices,
	ResourceUsers,
	ResourceSessions,
	ResourceLogs,
	ResourceWebDAV,
	ResourceMetrics,
	ResourceFeedback,
	ResourceDefault,
}
