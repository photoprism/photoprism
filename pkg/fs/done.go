package fs

type Status int8

const (
	Found     Status = 1
	Processed Status = 2
)

type Done map[string]Status

func (s Status) Exists() bool {
	return s > 0
}

func (s Status) Processed() bool {
	return s >= Processed
}
