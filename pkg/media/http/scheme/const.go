package scheme

// Type represents a URL scheme type.
type Type = string

const (
	File       Type = "file"
	Data       Type = "data"
	Http       Type = "http"
	Https      Type = "https"
	Websocket  Type = "wss"
	Unix       Type = "unix"
	HttpUnix   Type = "http+unix"
	Unixgram   Type = "unixgram"
	Unixpacket Type = "unixpacket"
)

var (
	HttpsData = []string{Https, Data}
	HttpsHttp = []string{Https, Http}
)
