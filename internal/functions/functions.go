package functions

import (
	"math"
)

// Function to safely convert from int64 to int32 due to Gorm2 changing count to return int64.
func SafeInt64to32(input int64) (result int32) {
	if input > math.MaxInt32 {
		result = math.MaxInt32
	} else if input < math.MinInt32 {
		result = math.MinInt32
	} else {
		result = int32(input)
	}
	return result
}

// Function to safely convert from int64 to int due to Gorm2 changing count to return int64.
func SafeInt64toint(input int64) (result int) {
	if input > math.MaxInt32 {
		result = math.MaxInt32
	} else if input < math.MinInt32 {
		result = math.MinInt32
	} else {
		result = int(input)
	}
	return result
}
