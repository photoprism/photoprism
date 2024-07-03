package unix

// Minute is one minute in seconds.
const Minute int64 = 60

// Hour is one hour in seconds.
const Hour = Minute * 60

// Day is one day in seconds.
const Day = Hour * 24

// Week is one week in seconds.
const Week = Day * 7

// Month is about one month in seconds.
const Month = Day * 31

// Year is 365 days in seconds.
const Year = Day * 365

// YearInt is Year specified as integer.
const YearInt = int(Year)
