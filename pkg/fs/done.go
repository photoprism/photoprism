package fs

type Status int8

const (
	Found     Status = 1
	Processed Status = 2
)

type Done map[string]Status

// Processed counts the number of processed files.
func (d Done) Processed() int {
	count := 0

	for _, s := range d {
		if s.Processed() {
			count++
		}
	}

	return count
}

func (s Status) Exists() bool {
	return s > 0
}

func (s Status) Processed() bool {
	return s >= Processed
}
