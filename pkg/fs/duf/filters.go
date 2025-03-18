package duf

var (
	all         = false
	hideDevices = ""
	hideFs      = ""
	hideMp      = ""
	onlyDevices = ""
	onlyFs      = ""
	onlyMp      = ""
)

type FilterValues map[string]struct{}

func NewFilterValues(s ...string) FilterValues {
	if len(s) == 0 {
		return make(FilterValues)
	} else if len(s) == 1 {
		return parseCommaSeparatedValues(s[0])
	}

	result := make(FilterValues, len(s))

	for i := range s {
		result[s[i]] = struct{}{}
	}

	return result
}

// FilterOptions contains all filters.
type FilterOptions struct {
	HiddenDevices FilterValues
	OnlyDevices   FilterValues

	HiddenFilesystems FilterValues
	OnlyFilesystems   FilterValues

	HiddenMountPoints FilterValues
	OnlyMountPoints   FilterValues
}
