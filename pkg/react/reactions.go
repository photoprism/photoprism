package react

var (
	Love      Emoji = "❤️"
	Like      Emoji = "👍"
	CatLove   Emoji = "😻"
	LoveIt    Emoji = "😍"
	InLove    Emoji = "🥰"
	Heart     Emoji = Love
	Cheers    Emoji = "🥂"
	Hot       Emoji = "🔥"
	Party     Emoji = "🎉"
	Birthday  Emoji = "🎂️"
	Sparkles  Emoji = "✨"
	Rainbow   Emoji = "🌈"
	Pride     Emoji = "🏳️‍🌈"
	SeeNoEvil Emoji = "🙈"
	Unknown   Emoji = ""
)

// Reactions specifies reaction emojis by name.
var Reactions = map[string]Emoji{
	"love":        Love,
	"+1":          Like,
	"cat-love":    CatLove,
	"love-it":     LoveIt,
	"in-love":     InLove,
	"heart":       Heart,
	"cheers":      Cheers,
	"hot":         Hot,
	"party":       Party,
	"birthday":    Birthday,
	"sparkles":    Sparkles,
	"rainbow":     Rainbow,
	"pride":       Pride,
	"see-no-evil": SeeNoEvil,
}

// Names specifies the reaction names by emoji.
var Names = map[Emoji]string{
	Love:      "love",
	Like:      "+1",
	CatLove:   "cat-love",
	LoveIt:    "love-it",
	InLove:    "in-love",
	Heart:     "heart",
	Cheers:    "cheers",
	Hot:       "hot",
	Party:     "party",
	Birthday:  "birthday",
	Sparkles:  "sparkles",
	Rainbow:   "rainbow",
	Pride:     "pride",
	SeeNoEvil: "see-no-evil",
}
