package query

import (
	"strings"
)

type CategoryLabel struct {
	Name  string
	Title string
}

func (s *Repo) CategoryLabels(limit, offset int) (results []CategoryLabel) {
	q := s.db.NewScope(nil).DB()

	q = q.Table("categories").
		Select("label_name AS name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Group("label_name").
		Limit(limit).Offset(offset)

	if err := q.Scan(&results).Error; err != nil {
		log.Errorf("categories: %s", err.Error())
		return results
	}

	for i, l := range results {
		results[i].Title = strings.Title(l.Name)
	}

	return results
}
