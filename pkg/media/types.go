package media

// Content is categorized and compared based on the following media types:
const (
	Unknown  Type = ""
	Image    Type = "image"
	Raw      Type = "raw"
	Live     Type = "live"
	Video    Type = "video"
	Animated Type = "animated"
	Audio    Type = "audio"
	Vector   Type = "vector"
	Document Type = "document"
	Sidecar  Type = "sidecar"
	Archive  Type = "archive"
)

// PriorityImage specifies the minimum priority for main media types, which can be indexed and displayed on their own,
// unlike archives or sidecar files that cannot be indexed or searched without a related main media file.
const (
	PriorityUnknown   = 0
	PrioritySidecar   = 1
	PriorityArchive   = 2
	PriorityImage     = 4
	PriorityMainMedia = PriorityImage
)

// Priorities maps media types to integer values that represent their relative importance.
type Priorities map[Type]int

// Priority assigns a relative priority value to the media type constants defined above.
var Priority = Priorities{
	Unknown:  PriorityUnknown, // 0
	Sidecar:  PrioritySidecar, // 1
	Archive:  PriorityArchive, // 2
	Image:    PriorityImage,   // 4
	Video:    8,
	Animated: 16,
	Audio:    16,
	Document: 16,
	Raw:      32,
	Vector:   32,
	Live:     64,
}
