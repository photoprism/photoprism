package video

// AvcProfiles contains common MPEG-4 AVC profiles together with their ContentType ID,
// maximum bitrate, resolution, and frame rate (if known):
// - https://dmnsgn.github.io/media-codecs/#AVC
// - https://en.wikipedia.org/wiki/Advanced_Video_Coding#Profiles
// - https://ott.dolby.com/codec_test/index.html
// - https://cconcolato.github.io/media-mime-support/
// - https://w3c.github.io/media-capabilities/
// - https://privacycheck.sec.lrz.de/active/fp_cpt/fp_can_play_type.html
// - https://chromium.googlesource.com/chromium/src.git/+/62.0.3178.1/content/browser/media/media_canplaytype_browsertest.cc
var AvcProfiles = CodecProfiles{
	{Codec: CodecAvc1, Profile: ProfileBaseline, Level: 30, Bitrate: 10000, ID: "avc1.66.30"}, // iOS friendly
	{Codec: CodecAvc1, Profile: ProfileBaseline, Level: 30, Bitrate: 10000, ID: "avc1.42001e"},
	{Codec: CodecAvc1, Profile: ProfileBaseline, Level: 31, Bitrate: 14000, ID: "avc1.42001f"},
	{Codec: CodecAvc1, Profile: ProfileMain, Level: 30, Bitrate: 10000, ID: "avc1.77.30"}, // iOS friendly
	{Codec: CodecAvc1, Profile: ProfileMain, Level: 30, Bitrate: 10000, ID: "avc1.4d001e"},
	{Codec: CodecAvc1, Profile: ProfileMain, Level: 31, Bitrate: 14000, ID: "avc1.4d001f"},
	{Codec: CodecAvc1, Profile: ProfileMain, Level: 40, Bitrate: 20000, ID: "avc1.4d0028"},
	{Codec: CodecAvc1, Profile: ProfileHigh, Level: 31, Bitrate: 17500, ID: "avc1.64001f"},
	{Codec: CodecAvc1, Profile: ProfileHigh, Level: 40, Bitrate: 25000, ID: "avc1.640028"},
	{Codec: CodecAvc1, Profile: ProfileHigh, Level: 41, Bitrate: 62500, ID: "avc1.640029"},
}
