package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// TestDB holds a test database connection.
type TestDB struct {
	DB       *sql.DB
	Driver   string
	DSN      string
}

// NewTestDB creates a new test database connection and runs migrations.
// It expects a TEST_DATABASE_URL environment variable.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatalf("failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres", driver)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDB{
		DB:     db,
		Driver: "pgx",
		DSN:    dsn,
	}
}

// Cleanup truncates all tables and closes the connection.
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()
	defer tdb.DB.Close()

	tables := []string{"refresh_tokens", "users"}
	for _, table := range tables {
		_, err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

// TruncateTable truncates a specific table.
func (tdb *TestDB) TruncateTable(t *testing.T, table string) {
	t.Helper()
	_, err := tdb.DB.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	if err != nil {
		t.Fatalf("failed to truncate table %s: %v", table, err)
	}
}
