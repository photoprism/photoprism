package acl

// Events that Roles can be granted Permission to listen to.
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
