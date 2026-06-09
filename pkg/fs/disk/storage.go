package disk

const (
	MB                     uint64 = 1024 * 1024 // MB defines one megabyte.
	DefaultStorageLowPct          = 1.0         // DefaultStorageLowPct defines free storage percent default.
	DefaultStorageLowBytes        = 100 * MB    // DefaultStorageLowBytes defines free storage bytes default.
)

// StorageLowPct is the percentage of total storage space below which it is considered critically low.
var StorageLowPct = DefaultStorageLowPct

// StorageLowBytes is the number of free bytes below which storage is considered critically low.
var StorageLowBytes = DefaultStorageLowBytes

// StorageLow reports free bytes on path and whether free space is critically low,
// i.e. below the StorageLowBytes floor or below StorageLowPct percent of total capacity.
// Setting StorageLowPct outside the open range (0, 100) disables the check entirely,
// including the StorageLowBytes floor.
func StorageLow(path string) (free uint64, low bool, err error) {
	if StorageLowPct <= 0 || StorageLowPct >= 100 {
		free, _, err = Free(path)
		return free, false, err
	}

	free, total, err := Free(path)

	if err != nil {
		return 0, false, err
	} else if total == 0 {
		return free, false, nil
	}

	low = free < StorageLowBytes || float64(free)*100 < StorageLowPct*float64(total)

	return free, low, nil
}
