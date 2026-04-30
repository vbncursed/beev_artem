package persistence

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // sql.Open("pgx", ...) driver
	"github.com/pressly/goose/v3"
)

// initTimeout bounds the total time a startup pass (connect + ping + apply
// migrations) is allowed to take. Without a deadline a stalled postgres would
// hang the boot indefinitely — the container would never get to the point
// where Docker / k8s could give up and restart it.
const initTimeout = 30 * time.Second

// embeddedMigrations are baked into the binary at compile time so the final
// alpine image needs no extra files. The directory layout is the standard
// goose convention: NNNNN_name.sql with `-- +goose Up/Down` blocks.
//
//go:embed migrations/*.sql
var embeddedMigrations embed.FS

type ResumeStorage struct {
	db *pgxpool.Pool
}

// Close releases all connections in the pool. Safe to call multiple times.
func (s *ResumeStorage) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// NewResumeStorage opens a pgx pool, verifies connectivity with a Ping, and
// applies any outstanding migrations through goose. Migrations replace the
// previous "CREATE TABLE IF NOT EXISTS at boot" pattern so future schema
// changes (add column, rename, drop FK, …) can be expressed as ordered SQL
// files instead of imperative Go code.
func NewResumeStorage(connString string) (*ResumeStorage, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// pgx v5 connects lazily — force a real handshake now so the timeout
	// applies to actual TCP/auth, not just config validation.
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if err := applyMigrations(ctx, connString); err != nil {
		db.Close()
		return nil, err
	}

	return &ResumeStorage{db: db}, nil
}

// applyMigrations opens a short-lived database/sql connection (goose's API
// requires *sql.DB) and runs UpContext. We use a separate connection rather
// than bridging the pgx pool because migrations run exactly once at boot —
// the extra round-trip cost is negligible and the code stays simpler.
func applyMigrations(ctx context.Context, connString string) error {
	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping migration db: %w", err)
	}

	goose.SetBaseFS(embeddedMigrations)
	// Per-service version tracking: every service in this monorepo shares
	// the same `hr` postgres database. The default goose table name
	// (`goose_db_version`) would collide across services — auth's v1 would
	// make vacancy think its own v1 had already been applied, the ALTER
	// in v2 would fire against a non-existent table, and the boot would
	// fail. Each service keeps its own progress in its own table.
	goose.SetTableName("resume_goose_db_version")
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
