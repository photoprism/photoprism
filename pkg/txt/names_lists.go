package txt

func init() {
	IsNameSuffix = make(map[string]bool, len(NameSuffixes))
	for _, n := range NameSuffixes {
		IsNameSuffix[n] = true
	}

	IsNameTitle = make(map[string]bool, len(NameTitles))
	for _, n := range NameTitles {
		IsNameTitle[n] = true
	}

}

var IsNameSuffix map[string]bool
var IsNameTitle map[string]bool

var NameSuffixes = []string{"esq", "esquire", "jr", "jnr", "sr", "snr", "2", "ii", "iii", "iv",
	"v", "clu", "chfc", "cfp", "md", "phd", "j.d.", "ll.m.", "m.d.", "d.o.", "d.c.",
	"p.c.", "ph.d."}

var NameTitles = []string{"mr", "mrs", "ms", "miss", "dr", "herr", "monsieur", "hr", "frau",
	"a v m", "admiraal", "admiral", "air cdre", "air commodore", "air marshal",
	"air vice marshal", "alderman", "alhaji", "ambassador", "baron", "barones",
	"brig", "brig gen", "brig general", "brigadier", "brigadier general",
	"brother", "canon", "capt", "captain", "cardinal", "cdr", "chief", "cik", "cmdr",
	"coach", "col", "colonel", "commandant", "commander", "commissioner",
	"commodore", "comte", "comtessa", "congressman", "conseiller", "consul",
	"conte", "contessa", "corporal", "councillor", "count", "countess",
	"crown prince", "crown princess", "dame", "datin", "dato", "datuk",
	"datuk seri", "deacon", "deaconess", "dean", "dhr", "dipl ing", "doctor",
	"dott", "dott sa", "dr ing", "dra", "drs", "embajador", "embajadora", "en",
	"encik", "eng", "eur ing", "exma sra", "exmo sr", "f o", "father",
	"first lieutient", "first officer", "flt lieut", "flying officer", "fr",
	"frau", "fraulein", "fru", "gen", "generaal", "general", "governor", "graaf",
	"gravin", "group captain", "grp capt", "h e dr", "h h", "h m", "h r h", "hajah",
	"haji", "hajim", "her highness", "her majesty", "herr", "high chief",
	"his highness", "his holiness", "his majesty", "hon", "hr", "hra", "ing", "ir",
	"jonkheer", "judge", "justice", "khun ying", "kolonel", "lady", "lcda", "lic",
	"lieut", "lieut cdr", "lieut col", "lieut gen", "lord", "m", "m l", "m r",
	"madame", "mademoiselle", "maj gen", "major", "master", "mevrouw", "miss",
	"mlle", "mme", "monsieur", "monsignor", "mstr", "nti", "pastor",
	"president", "prince", "princess", "princesse", "prinses", "prof",
	"prof sir", "professor", "puan", "puan sri", "rabbi", "rear admiral", "rev",
	"rev canon", "rev dr", "rev mother", "reverend", "rva", "senator", "sergeant",
	"sheikh", "sheikha", "sig", "sig na", "sig ra", "sir", "sister", "sqn ldr", "sr",
	"sr d", "sra", "srta", "sultan", "tan sri", "tan sri dato", "tengku", "teuku",
	"than puying", "the hon dr", "the hon justice", "the hon miss", "the hon mr",
	"the hon mrs", "the hon ms", "the hon sir", "the very rev", "toh puan", "tun",
	"vice admiral", "viscount", "viscountess", "wg cdr"}
