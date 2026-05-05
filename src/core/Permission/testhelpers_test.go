package permission

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Import goose for migrations
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var (
	testDB   *gorm.DB
	dbOnce   sync.Once
	dbDSN    string
)

// setupTestContainer creates a PostgreSQL container and runs migrations.
func setupTestContainer(t *testing.T) *gorm.DB {
	t.Helper()

	dbOnce.Do(func() {
		ctx := context.Background()

		// Start PostgreSQL container
		pgContainer, err := postgres.Run(ctx,
			"postgres:15-alpine",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("testuser"),
			postgres.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			t.Fatalf("failed to start postgres container: %v", err)
		}

		// Get connection string
		connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			t.Fatalf("failed to get connection string: %v", err)
		}
		dbDSN = connStr

		// Run migrations using goose
		db, err := sql.Open("postgres", dbDSN)
		if err != nil {
			t.Fatalf("failed to open db for migration: %v", err)
		}
		defer db.Close()

		// Setup goose
		if err := goose.SetDialect("postgres"); err != nil {
			t.Fatalf("failed to set goose dialect: %v", err)
		}

		// Get migration directory
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("failed to get current file path")
		}
		projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filename))))
		migrationDir := filepath.Join(projectRoot, "migration")

		if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
			t.Fatalf("migration directory not found at %s: %v", migrationDir, err)
		}

		// Run goose up
		if err := goose.Up(db, migrationDir); err != nil {
			t.Fatalf("failed to run migrations: %v", err)
		}

		// Connect with GORM
		testDB, err = gorm.Open(gormpostgres.Open(dbDSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			t.Fatalf("failed to connect to db with gorm: %v", err)
		}
	})

	// Truncate tables before each test for isolation
	t.Cleanup(func() {
		if testDB != nil {
			truncateTables(t, testDB)
		}
	})

	return testDB
}

// truncateTables truncates all tables for test isolation
func truncateTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	db.Exec("SET CONSTRAINTS ALL DEFERRED")

	var tables []string
	db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)

	for _, table := range tables {
		if table != "schema_migrations" {
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %q CASCADE", table))
		}
	}
}

// RED phase: Test that setupTestContainer works
func TestSetupTestContainer_ReturnsDB(t *testing.T) {
	db := setupTestContainer(t)
	assert.NotNil(t, db, "expected db to be non-nil")
}

func TestSetupTestContainer_CanQueryPermissions(t *testing.T) {
	db := setupTestContainer(t)

	// Try to query permissions table (should exist after migrations)
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM permissions").Scan(&count).Error
	assert.NoError(t, err, "failed to query permissions table")
}
