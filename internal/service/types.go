package service

// Type represents a remote service type, e.g. WebDAV.
type Type = string

// Identifiers for common remote services.
const (
	WebDAV    Type = "webdav"
	Facebook  Type = "facebook"
	Twitter   Type = "twitter"
	Flickr    Type = "flickr"
	Instagram Type = "instagram"
	Telegram  Type = "telegram"
	WhatsApp  Type = "whatsapp"
	GPhotos   Type = "gphotos"
	GDrive    Type = "gdrive"
	OneDrive  Type = "onedrive"
)
