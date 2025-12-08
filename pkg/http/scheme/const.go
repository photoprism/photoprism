package scheme

// Type represents a URL scheme type.
type Type = string

const (
	// File scheme.
	File Type = "file"
	// Data scheme.
	Data Type = "data"
	// Base64 scheme.
	Base64 Type = "base64"
	// Http scheme.
	Http Type = "http"
	// Https scheme.
	Https Type = "https"
	// Websocket scheme (secure).
	Websocket Type = "wss"
	// Unix scheme.
	Unix Type = "unix"
	// HttpUnix scheme.
	HttpUnix Type = "http+unix"
	// Unixgram scheme.
	Unixgram Type = "unixgram"
	// Unixpacket scheme.
	Unixpacket Type = "unixpacket"
)

var (
	// HttpsData lists allowed schemes (https, data).
	HttpsData = []string{Https, Data}
	// HttpsHttp lists allowed schemes (https, http).
	HttpsHttp = []string{Https, Http}
)
