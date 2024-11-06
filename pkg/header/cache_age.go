package header

// CacheMinute is one minute in seconds.
const CacheMinute int64 = 60

// CacheHour is one hour in seconds.
const CacheHour = CacheMinute * 60

// CacheDay is one day in seconds.
const CacheDay = CacheHour * 24

// CacheWeek is one week in seconds.
const CacheWeek = CacheDay * 7

// CacheMonth is about one month in seconds.
const CacheMonth = CacheDay * 31

// CacheYear is 365 days in seconds.
const CacheYear = CacheDay * 365

// CacheYearInt is CacheYear specified as integer.
const CacheYearInt = int(CacheYear)
