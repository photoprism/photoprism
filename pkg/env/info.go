package env

// Info returns a new Resources instance.
func Info() (resources Resources) {
	resources = Resources{}
	resources.Update()

	return resources
}
