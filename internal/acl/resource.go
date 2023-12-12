package acl

import "strings"

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
	ResourcePassword  Resource = "password"
	ResourceServices  Resource = "services"
	ResourceUsers     Resource = "users"
	ResourceLogs      Resource = "logs"
	ResourceWebDAV    Resource = "webdav"
	ResourceMetrics   Resource = "metrics"
	ResourceFeedback  Resource = "feedback"
	ResourceDefault   Resource = "default"
)

// Resource represents a resource for which roles can be granted Permission.
type Resource string

// String returns the type as string.
func (r Resource) String() string {
	if r == "" {
		return "default"
	}

	return string(r)
}

// LogId returns an identifier string for use in log messages.
func (r Resource) LogId() string {
	return r.String()
}

// Equal checks if the type matches.
func (r Resource) Equal(s string) bool {
	return strings.EqualFold(s, r.String())
}

// NotEqual checks if the type is different.
func (r Resource) NotEqual(s string) bool {
	return !r.Equal(s)
}
