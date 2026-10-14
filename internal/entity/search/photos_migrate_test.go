package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
)

// TestPhoto_Migrate checks that migrating the entity tables after the search result type has been parsed
// leaves the photos table as the entity defines it.
func TestPhoto_Migrate(t *testing.T) {
	// Parsing the result type registers its file relation on the shared file schema.
	stmt := &gorm.Statement{DB: entity.Db()}
	require.NoError(t, stmt.Parse(&Photo{}))

	require.NoError(t, entity.Tables{"files": &entity.File{}}.Migrate(entity.Db(), migrate.Opt(true, false, nil)))

	found, err := entity.DbHasColumn(entity.Db(), "photos", "composite_id")
	require.NoError(t, err)
	assert.False(t, found, "photos table has no search result columns")
}
