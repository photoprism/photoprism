package photoprism

// TimeZone returns the time zone where the photo was taken.
func (m *MediaFile) TimeZone() string {
	data := m.MetaData()

	return data.TimeZone
}
