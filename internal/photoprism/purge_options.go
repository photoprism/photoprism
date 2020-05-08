package photoprism

type PurgeOptions struct {
	Path   string
	Ignore map[string]bool
	Dry    bool
	Hard   bool
}
