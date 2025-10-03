package provisioner

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestGenerateCredentials_StabilityAndBudgets(t *testing.T) {
	c := config.NewConfig(config.CliTestContext())
	// Fix the cluster UUID via options to ensure determinism.
	c.Options().ClusterUUID = "11111111-1111-4111-8111-111111111111"

	db1, user1, pass1 := GenerateCredentials(c, "11111111-1111-4111-8111-111111111111", "pp-node-01")
	db2, user2, pass2 := GenerateCredentials(c, "11111111-1111-4111-8111-111111111111", "pp-node-01")

	// Names stable; password random.
	assert.Equal(t, db1, db2)
	assert.Equal(t, user1, user2)
	assert.NotEqual(t, pass1, pass2)

	// Budgets and patterns.
	assert.LessOrEqual(t, len(user1), 32)
	assert.LessOrEqual(t, len(db1), 64)
	assert.Contains(t, db1, "photoprism_")
	assert.Contains(t, user1, "photoprism_")
}

func TestGenerateCredentials_DifferentPortal(t *testing.T) {
	c1 := config.NewConfig(config.CliTestContext())
	c2 := config.NewConfig(config.CliTestContext())
	c1.Options().ClusterUUID = "11111111-1111-4111-8111-111111111111"
	c2.Options().ClusterUUID = "22222222-2222-4222-8222-222222222222"

	db1, user1, _ := GenerateCredentials(c1, "11111111-1111-4111-8111-111111111111", "pp-node-01")
	db2, user2, _ := GenerateCredentials(c2, "11111111-1111-4111-1111-111111111111", "pp-node-01")

	assert.NotEqual(t, db1, db2)
	assert.NotEqual(t, user1, user2)
}

func TestGenerateCredentials_Truncation(t *testing.T) {
	c := config.NewConfig(config.CliTestContext())
	c.Options().ClusterUUID = "11111111-1111-4111-8111-111111111111"
	longName := "this-is-a-very-very-long-node-name-that-should-be-truncated-to-fit-username-and-db-budgets"
	db, user, _ := GenerateCredentials(c, "11111111-1111-4111-8111-111111111111", longName)

	assert.LessOrEqual(t, len(user), 32)
	assert.LessOrEqual(t, len(db), 64)
}

func TestBuildDSN(t *testing.T) {
	dsn := BuildDSN("mysql", "mariadb", 3306, "user", "pass", "dbname")
	assert.Contains(t, dsn, "user:pass@tcp(mariadb:3306)/dbname")
	assert.Contains(t, dsn, "charset=utf8mb4")
	assert.Contains(t, dsn, "parseTime=true")
}

func TestHmacBase32_LowercaseDeterministic(t *testing.T) {
	a := hmacBase32("k1", "data")
	b := hmacBase32("k1", "data")
	c := hmacBase32("k1", "other")

	assert.Equal(t, a, b, "same key/data should produce identical digest")
	assert.NotEqual(t, a, c, "different data should change the digest")
	assert.NotZero(t, len(a))
	assert.Equal(t, strings.ToLower(a), a, "digest must be lowercase")
	for _, ch := range a {
		assert.Contains(t, "abcdefghijklmnopqrstuvwxyz234567", string(ch))
	}
}

func TestGetCredentials_SqliteRejected(t *testing.T) {
	ctx := context.Background()
	c := config.NewConfig(config.CliTestContext())
	origDriver := DatabaseDriver
	DatabaseDriver = config.SQLite3
	t.Cleanup(func() { DatabaseDriver = origDriver })

	_, _, err := GetCredentials(ctx, c, "11111111-1111-4111-8111-111111111111", "pp-node-01", false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "database must be MySQL/MariaDB")
	}
}
