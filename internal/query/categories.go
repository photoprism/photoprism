package query

import "github.com/photoprism/photoprism/pkg/txt"

type CategoryLabel struct {
	Name  string
	Title string
}

func CategoryLabels(limit, offset int) (results []CategoryLabel) {
	s := Db().NewScope(nil).DB()

	s = s.Table("categories").
		Select("label_name AS name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Group("label_name").
		Limit(limit).Offset(offset)

	if err := s.Scan(&results).Error; err != nil {
		log.Errorf("categories: %s", err.Error())
		return results
	}

	for i, l := range results {
		results[i].Title = txt.Title(l.Name)
	}

	return results
}
