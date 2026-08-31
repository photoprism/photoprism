package report

import "strconv"

// DistancePrecision is the number of decimals a distance is reported with. Face thresholds are
// stated to three, so a report that shows fewer cannot be compared against them.
const DistancePrecision = 3

// Distance formats an embedding distance at the precision its thresholds are stated in, so a
// reported value and a configured one can be read against each other.
func Distance(f float64) string {
	return strconv.FormatFloat(f, 'f', DistancePrecision, 64)
}
