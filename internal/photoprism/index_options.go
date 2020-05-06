package photoprism

type IndexOptions struct {
	Rescan         bool
	Convert        bool
}

func (o *IndexOptions) SkipUnchanged() bool {
	return !o.Rescan
}

// IndexOptionsAll returns new index options with all options set to true.
func IndexOptionsAll() IndexOptions {
	result := IndexOptions{
		Rescan:         true,
		Convert:        true,
	}

	return result
}

// IndexOptionsNone returns new index options with all options set to false.
func IndexOptionsNone() IndexOptions {
	result := IndexOptions{}

	return result
}
