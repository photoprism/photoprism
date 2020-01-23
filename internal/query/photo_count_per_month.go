package query

// PhotoCountResult contains photos per month
type PhotoCountResult struct {
	PhotoYear  int
	PhotoMonth int
	Count      int
}

// GetPhotoCountPerMonth couts photos per month and year
func (s *Repo) GetPhotoCountPerMonth() (results []PhotoCountResult, err error) {
	q := s.db.NewScope(nil).DB()

	q = q.Table("photos").
		Select("photos.photo_year, photos.photo_month, COUNT(*) AS count").
		Group("photos.photo_year, photos.photo_month").
		Order("photos.photo_year DESC, photos.photo_month DESC")

	if result := q.Scan(&results); result.Error != nil {
		return results, result.Error
	}

	return results, nil
}
