package video

// Profile represents a video codec profile name,
// see https://en.wikipedia.org/wiki/Advanced_Video_Coding#Profiles.
type Profile = string

const (
	// ProfileBaseline indicates H.264 Baseline profile.
	ProfileBaseline Profile = "Baseline"
	// ProfileMain indicates H.264 Main profile.
	ProfileMain Profile = "Main"
	// ProfileHigh indicates H.264 High profile.
	ProfileHigh Profile = "High"
)

// CodecProfile represents a codec subtype with its standardized ID,
// maximum bitrate, resolution, and frame rate (if known).
type CodecProfile struct {
	Codec      Codec
	Profile    string
	Level      int
	SubLevel   int
	Bitrate    int
	Resolution int
	Framerate  int
	ID         string
}

// CodecProfiles represents a set of codec subtypes.
type CodecProfiles []CodecProfile
