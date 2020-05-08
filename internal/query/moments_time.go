package query

// MomentsTimeResult contains photo counts per month and year
type MomentsTimeResult struct {
	PhotoYear  int
	PhotoMonth int
	Count      int
}

// GetMomentsTime counts photos per month and year
func GetMomentsTime() (results []MomentsTimeResult, err error) {
	s := UnscopedDb()

	s = s.Table("photos").
		Where("deleted_at IS NULL").
		Select("photos.photo_year, photos.photo_month, COUNT(*) AS count").
		Group("photos.photo_year, photos.photo_month").
		Order("photos.photo_year DESC, photos.photo_month DESC")

	if result := s.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
