package header

type RobotsRule = string

// Robots controls how pages are indexed and crawled by search engines:
// https://developers.google.com/search/docs/crawling-indexing/robots-meta-tag#xrobotstag
const (
	Robots = "X-Robots-Tag"
)

// Standard Robots header values.
const (
	RobotsAll      RobotsRule = "all"
	RobotsNone     RobotsRule = "noindex, nofollow"
	RobotsNoIndex  RobotsRule = "noindex"
	RobotsNoFollow RobotsRule = "nofollow"
	RobotsNoImages RobotsRule = "noimageindex"
)
