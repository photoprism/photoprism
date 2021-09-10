package photoprism

const (
	MotionPhotoSamsung = "MotionPhoto_Data"
	MotionPhotoGoogle  = "MotionPhoto"
)

func (m *MediaFile) IsMotionPhoto() bool {
	if m.MetaData().MotionPhoto {
		// Google MotionPhoto v1
		return true
	} else if m.MetaData().EmbeddedVideoType == MotionPhotoSamsung {
		// Samsung MotionPhoto
		return true
	}

	return false
}
