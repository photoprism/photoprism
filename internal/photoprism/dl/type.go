package dl

// Type of response you want
type Type int

const (
	// TypeAny single or playlist (default)
	TypeAny Type = iota
	// TypeSingle single track, file etc
	TypeSingle
	// TypePlaylist playlist with multiple tracks, files etc
	TypePlaylist
	// TypeChannel channel containing one or more playlists, which will be flattened
	TypeChannel
)

// TypeFromString translates string identifiers to download Type values.
var TypeFromString = map[string]Type{
	"any":      TypeAny,
	"single":   TypeSingle,
	"playlist": TypePlaylist,
	"channel":  TypeChannel,
}
