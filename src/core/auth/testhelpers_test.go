package authentication

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

// createTestRole creates a test role for testing.
func createTestRole(t *testing.T, db *gorm.DB) *model.Role {
	t.Helper()
	role := &model.Role{
		ID:   uuid.New(),
		Name: "Test Role",
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}
	return role
}

// createTestUser creates a test user with a role for testing.
func createTestUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()

	role := createTestRole(t, db)

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

// createTestPermission creates a test permission for testing.
func createTestPermission(t *testing.T, db *gorm.DB, code string) *model.Permission {
	t.Helper()
	perm := &model.Permission{
		ID:          code,
		Name:        code,
		Description: "Test permission",
	}
	if err := db.Create(perm).Error; err != nil {
		t.Fatalf("failed to create test permission: %v", err)
	}
	return perm
}

// assignRolePermission assigns a permission to a role.
func assignRolePermission(t *testing.T, db *gorm.DB, roleID uuid.UUID, permissionID string) {
	t.Helper()
	rp := &model.RolePermission{
		ID:           uuid.New(),
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	if err := db.Create(rp).Error; err != nil {
		t.Fatalf("failed to assign role permission: %v", err)
	}
}

// assignUserPermission assigns a permission directly to a user.
func assignUserPermission(t *testing.T, db *gorm.DB, userID uuid.UUID, permissionID string) {
	t.Helper()
	up := &model.UserPermission{
		ID:           uuid.New(),
		UserID:       userID,
		PermissionID: permissionID,
	}
	if err := db.Create(up).Error; err != nil {
		t.Fatalf("failed to assign user permission: %v", err)
	}
}

// createOAuthState creates and saves a test OAuth state for testing.
func createOAuthState(t *testing.T, db *gorm.DB, provider string) string {
	t.Helper()
	state := uuid.New().String()
	oauthState := model.OAuthState{
		ID:        uuid.New(),
		State:     state,
		Provider:  provider,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := db.Create(&oauthState).Error; err != nil {
		t.Fatalf("failed to create oauth state: %v", err)
	}
	return state
}

// createTestSession creates a test session for a user.
func createTestSession(t *testing.T, db *gorm.DB, userID uuid.UUID) *model.Session {
	t.Helper()
	session := &model.Session{
		ID:               uuid.New(),
		UserID:           userID,
		Provider:         "local",
		RefreshTokenHash: fmt.Sprintf("hash-%s", uuid.New().String()),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Ip:               "127.0.0.1",
		UserAgent:        "test-agent",
	}
	if err := db.Create(session).Error; err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}
	return session
}

// createTestToken creates a test JWT token for testing.
func createTestToken(t *testing.T, db *gorm.DB) string {
	t.Helper()
	return uuid.New().String()
}

// RED phase: Test that setupTestContainer works
func TestSetupTestContainer_ReturnsDB(t *testing.T) {
	db := setupTestContainer(t)
	assert.NotNil(t, db, "expected db to be non-nil")
}

func TestSetupTestContainer_CanQuerySessions(t *testing.T) {
	db := setupTestContainer(t)

	// Try to query sessions table (should exist after migrations)
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM sessions").Scan(&count).Error
	assert.NoError(t, err, "failed to query sessions table")
}

// RED phase: Test OAuth state helper
func TestCreateOAuthState_ReturnsValidState(t *testing.T) {
	db := setupTestContainer(t)

	state := createOAuthState(t, db, "google")

	assert.NotEmpty(t, state, "expected OAuth state to be non-empty")
}

// RED phase: Test token helper
func TestCreateTestToken_ReturnsToken(t *testing.T) {
	db := setupTestContainer(t)

	token := createTestToken(t, db)

	assert.NotEmpty(t, token, "expected token to be non-empty")
}
