package entity

type PhotoKeywordMap map[string]PhotoKeyword

var PhotoKeywordFixtures = PhotoKeywordMap{
	"1": {
		PhotoID:   1000004,
		KeywordID: 1000000,
	},
	"2": {
		PhotoID:   1000001,
		KeywordID: 1000001,
	},
	"3": {
		PhotoID:   1000000,
		KeywordID: 1000000,
	},
	"4": {
		PhotoID:   1000003,
		KeywordID: 1000000,
	},
	"5": {
		PhotoID:   1000003,
		KeywordID: 1000003,
	},
	"6": {
		PhotoID:   1000023,
		KeywordID: 1000003,
	},
	"7": {
		PhotoID:   1000027,
		KeywordID: 1000004,
	},
	"8": {
		PhotoID:   1000030,
		KeywordID: 1000007,
	},
	"9": {
		PhotoID:   10000029,
		KeywordID: 1000005,
	},
	"10": {
		PhotoID:   1000031,
		KeywordID: 1000006,
	},
	"11": {
		PhotoID:   1000032,
		KeywordID: 1000008,
	},
	"12": {
		PhotoID:   1000033,
		KeywordID: 1000009,
	},
	"13": {
		PhotoID:   1000034,
		KeywordID: 10000010,
	},
	"14": {
		PhotoID:   1000035,
		KeywordID: 10000011,
	},
	"15": {
		PhotoID:   1000036,
		KeywordID: 10000012,
	},
	"16": {
		PhotoID:   1000037,
		KeywordID: 10000013,
	},
	"17": {
		PhotoID:   1000038,
		KeywordID: 10000014,
	},
	"18": {
		PhotoID:   1000039,
		KeywordID: 10000015,
	},
	"19": {
		PhotoID:   1000040,
		KeywordID: 10000016,
	},
	"20": {
		PhotoID:   1000041,
		KeywordID: 10000017,
	},
	"21": {
		PhotoID:   1000042,
		KeywordID: 10000018,
	},
	"22": {
		PhotoID:   1000043,
		KeywordID: 10000019,
	},
	"23": {
		PhotoID:   1000044,
		KeywordID: 10000020,
	},
	"24": {
		PhotoID:   1000045,
		KeywordID: 10000021,
	},
	"25": {
		PhotoID:   1000046,
		KeywordID: 10000022,
	},
	"26": {
		PhotoID:   1000047,
		KeywordID: 10000023,
	},
	"27": {
		PhotoID:   1000048,
		KeywordID: 10000024,
	},
	"28": {
		PhotoID:   1000049,
		KeywordID: 10000025,
	},
	"29": {
		PhotoID:   1000045,
		KeywordID: 10000018,
	},
	"30": {
		PhotoID:   1000045,
		KeywordID: 10000018,
	},
	"31": {
		PhotoID:   1000036,
		KeywordID: 10000015,
	},
	"32": {
		PhotoID:   1000036,
		KeywordID: 1000001,
	},
}

// CreatePhotoKeywordFixtures inserts known entities into the database for testing.
func CreatePhotoKeywordFixtures() {
	for _, entity := range PhotoKeywordFixtures {
		Db().Create(&entity)
	}
}
