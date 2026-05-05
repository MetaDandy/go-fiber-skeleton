package role

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

	"github.com/MetaDandy/go-fiber-skeleton/config/seed"
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
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("failed to get current file path")
		}
		// Go up 3 directories: role -> core -> src -> project root
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

		// Seed permissions (19 permissions)
		seed.Seeder(testDB)
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
// Preserves 'permissions' table (seeded data that persists)
func truncateTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Disable foreign key checks temporarily
	db.Exec("SET CONSTRAINTS ALL DEFERRED")

	// Get all table names
	var tables []string
	db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Scan(&tables)

	for _, table := range tables {
		// Skip schema_migrations and permissions (seeded data)
		if table != "schema_migrations" && table != "permissions" {
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %q CASCADE", table))
		}
	}
}

// createRoleHierarchy creates a 3-level role hierarchy for testing.
// Returns roles in order: [parent, child, grandchild].
func createRoleHierarchy(t *testing.T, db *gorm.DB) []model.Role {
	t.Helper()

	parentID := uuid.New()
	parent := model.Role{
		ID:          parentID,
		Name:        "Parent Role",
		Description: "Top-level role",
		RoleID:      nil,
	}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatalf("failed to create parent role: %v", err)
	}

	childID := uuid.New()
	child := model.Role{
		ID:          childID,
		Name:        "Child Role",
		Description: "Middle-level role",
		RoleID:      &parentID,
	}
	if err := db.Create(&child).Error; err != nil {
		t.Fatalf("failed to create child role: %v", err)
	}

	grandchildID := uuid.New()
	grandchild := model.Role{
		ID:          grandchildID,
		Name:        "Grandchild Role",
		Description: "Bottom-level role",
		RoleID:      &childID,
	}
	if err := db.Create(&grandchild).Error; err != nil {
		t.Fatalf("failed to create grandchild role: %v", err)
	}

	return []model.Role{parent, child, grandchild}
}

// RED phase: Test that setupTestContainer works
func TestSetupTestContainer_ReturnsDB(t *testing.T) {
	db := setupTestContainer(t)
	assert.NotNil(t, db, "expected db to be non-nil")
}

// RED phase: Test createRoleHierarchy helper
func TestCreateRoleHierarchy_CreatesThreeLevels(t *testing.T) {
	db := setupTestContainer(t)

	roles := createRoleHierarchy(t, db)

	assert.Equal(t, 3, len(roles), "expected 3 roles in hierarchy")

	// Verify they have the correct parent-child relationships
	for i, role := range roles {
		assert.NotEmpty(t, role.ID, "role ID should not be empty")

		if i > 0 {
			// Child and grandchild should have a parent
			assert.NotNil(t, role.RoleID, "role at index %d should have a parent", i)
		} else {
			// Parent should have no parent
			assert.Nil(t, role.RoleID, "parent role should have no parent")
		}
	}
}
