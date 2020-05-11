package entity

func CreateViews() {
	labelCounts := `CREATE OR REPLACE VIEW label_counts AS
		SELECT label_id, SUM(photo_count) AS photo_count FROM (
		(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
		            JOIN photos_labels pl ON pl.label_id = l.id
		            JOIN photos ph ON pl.photo_id = ph.id
					WHERE pl.uncertainty < 100
					AND ph.photo_quality >= 0
					AND ph.photo_private = 0
					AND ph.deleted_at IS NULL GROUP BY l.id)
		UNION ALL
		(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
		            JOIN categories c ON c.category_id = l.id
		            JOIN photos_labels pl ON pl.label_id = c.label_id
		            JOIN photos ph ON pl.photo_id = ph.id
					WHERE pl.uncertainty < 100
					AND ph.photo_quality >= 0
					AND ph.photo_private = 0
					AND ph.deleted_at IS NULL GROUP BY l.id)) counts GROUP BY label_id`

	if err := UnscopedDb().Exec(labelCounts).Error; err != nil {
		log.Error(err)
	}
}
