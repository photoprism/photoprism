package api

import "strconv"

func ParseUint(s string) uint {
	result, err := strconv.ParseUint(s, 10, 32)

	if err != nil {
		log.Warnf("api: %s", err)
	}

	return uint(result)
}
