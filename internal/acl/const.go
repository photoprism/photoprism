package acl

// Roles that can be granted Permissions to use a Resource.
const (
	RoleDefault Role = "default"
	RoleUser    Role = "user"
	RoleViewer  Role = "viewer"
	RoleGuest   Role = "guest"
	RoleAdmin   Role = "admin"
	RoleVisitor Role = "visitor"
	RoleClient  Role = "client"
	RoleNone    Role = ""
)

// Permissions to use a Resource that can be granted to a Role.
const (
	FullAccess      Permission = "full_access"
	AccessShared    Permission = "access_shared"
	AccessLibrary   Permission = "access_library"
	AccessPrivate   Permission = "access_private"
	AccessOwn       Permission = "access_own"
	AccessAll       Permission = "access_all"
	ActionSearch    Permission = "search"
	ActionView      Permission = "view"
	ActionUpload    Permission = "upload"
	ActionCreate    Permission = "create"
	ActionUpdate    Permission = "update"
	ActionDownload  Permission = "download"
	ActionShare     Permission = "share"
	ActionDelete    Permission = "delete"
	ActionRate      Permission = "rate"
	ActionReact     Permission = "react"
	ActionSubscribe Permission = "subscribe"
	ActionManage    Permission = "manage"
	ActionManageOwn Permission = "manage_own"
)

// A Role can be given Permission to use a Resource.
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

// Events for which a Role can be granted the ActionSubscribe Permission.
const (
	ChannelUser      Resource = "user"
	ChannelSession   Resource = "session"
	ChannelAudit     Resource = "audit"
	ChannelLog       Resource = "log"
	ChannelNotify    Resource = "notify"
	ChannelIndex     Resource = "index"
	ChannelUpload    Resource = "upload"
	ChannelImport    Resource = "import"
	ChannelConfig    Resource = "config"
	ChannelCount     Resource = "count"
	ChannelPhotos    Resource = "photos"
	ChannelCameras   Resource = "cameras"
	ChannelLenses    Resource = "lenses"
	ChannelCountries Resource = "countries"
	ChannelAlbums    Resource = "albums"
	ChannelLabels    Resource = "labels"
	ChannelSubjects  Resource = "subjects"
	ChannelPeople    Resource = "people"
	ChannelSync      Resource = "sync"
)
