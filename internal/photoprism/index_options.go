package photoprism

type IndexOptions struct {
	Path    string
	Rescan  bool
	Convert bool
	Single  bool
}

func (o *IndexOptions) SkipUnchanged() bool {
	return !o.Rescan
}

// IndexOptionsAll returns new index options with all options set to true.
func IndexOptionsAll() IndexOptions {
	result := IndexOptions{
		Path:    "/",
		Rescan:  true,
		Convert: true,
		Single:  false,
	}

	return result
}

// IndexOptionsSingle returns new index options for unstacked, single files.
func IndexOptionsSingle() IndexOptions {
	result := IndexOptions{
		Path:    "/",
		Rescan:  true,
		Convert: true,
		Single:  true,
	}

	return result
}

// IndexOptionsNone returns new index options with all options set to false.
func IndexOptionsNone() IndexOptions {
	result := IndexOptions{}

	return result
}
