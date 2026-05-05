package user

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

	"github.com/MetaDandy/go-fiber-skeleton/src/model"
	"github.com/google/uuid"
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
// Uses sync.Once to ensure it runs only once per test session.
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

		// Get migration directory - use runtime to get current file path
		// We're at src/core/user/testhelpers_test.go
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("failed to get current file path")
		}
		// Go up 3 directories: user -> core -> src -> project root
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

		// We don't terminate the container here - it lives for the test session
		// Container will be removed when test process exits
		// No cleanup needed for container - testcontainers handles it on process exit
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
	
	// Disable foreign key checks temporarily
	db.Exec("SET CONSTRAINTS ALL DEFERRED")
	
	// Get all table names
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

	// Verify we can ping the database
	sqlDB, err := db.DB()
	assert.NoError(t, err, "failed to get sql.DB")

	err = sqlDB.Ping()
	assert.NoError(t, err, "failed to ping database")
}

func TestSetupTestContainer_ExecutesMigrations(t *testing.T) {
	db := setupTestContainer(t)

	// Check that migrations ran by trying to query a table that should exist
	// After migration, the users table should exist
	type Result struct {
		Exists bool
	}
	var result Result

	// Query to check if users table exists
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users') as exists").Scan(&result).Error
	assert.NoError(t, err, "failed to check if users table exists")
	assert.True(t, result.Exists, "expected users table to exist after migrations")
}

// Helper to create a test user with a default role
func createTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()
	
	// Create a default role first to satisfy foreign key constraint
	role := &model.Role{
		ID:   uuid.New(),
		Name: "Test Role",
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}
	
	pwd := "hashedpassword"
	user := &model.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     fmt.Sprintf("test-%s@example.com", uuid.New().String()),
		Password:  &pwd,
		RoleID:    role.ID,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}
