package acl

type Resource string

const (
	ResourceDefault       Resource = "*"
	ResourceConfig        Resource = "config"
	ResourceConfigOptions Resource = "config_options"
	ResourceSettings      Resource = "settings"
	ResourceLogs          Resource = "logs"
	ResourceAccounts      Resource = "accounts"
	ResourceAlbums        Resource = "albums"
	ResourceCameras       Resource = "cameras"
	ResourceCategories    Resource = "categories"
	ResourceCountries     Resource = "countries"
	ResourceFiles         Resource = "files"
	ResourceFolders       Resource = "folders"
	ResourceLabels        Resource = "labels"
	ResourceLenses        Resource = "lenses"
	ResourceLinks         Resource = "links"
	ResourceGeo           Resource = "geo"
	ResourcePasswords     Resource = "passwords"
	ResourcePeople        Resource = "people"
	ResourcePhotos        Resource = "photos"
	ResourcePlaces        Resource = "places"
	ResourceFeedback      Resource = "feedback"
)
