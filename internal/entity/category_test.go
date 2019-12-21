package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategory_TableName(t *testing.T) {
	label := &Label{LabelSlug: "cute-kitten", LabelName: " Cute Kitten"}
	category := &Category{LabelID: 1, CategoryID: 1, Label: label, Category: label}
	tableName := category.TableName()

	assert.Equal(t, "categories", tableName)
}
