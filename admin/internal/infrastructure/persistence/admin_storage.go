// Package persistence implements admin's storage port via pgxpool.
// All queries are READ-ONLY against the shared `hr` database — admin is the
// only service that crosses per-service table ownership; documented as a
// deliberate operational exception in admin/README.md so adding aggregate
// counters doesn't require a new RPC in every domain service.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// initTimeout bounds the total time a startup pass (parse + connect + ping)
// is allowed to take. Without a deadline a stalled postgres would hang the
// boot indefinitely — the container would never get to the point where Docker
// could give up and restart it.
const initTimeout = 30 * time.Second

// AdminStorage holds a pgxpool connected to the shared `hr` database. Closed
// via Close() during graceful shutdown.
type AdminStorage struct {
	db *pgxpool.Pool
}

// NewAdminStorage initialises the pool and Pings to fail fast if the DB is
// unreachable. The pool is intentionally small — admin's workload is bursty
// dashboard reads, not steady throughput.
func NewAdminStorage(connString string) (*AdminStorage, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}
	cfg.MaxConns = 4

	ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}
	return &AdminStorage{db: pool}, nil
}

// Close releases all connections in the pool. Safe to call multiple times.
func (s *AdminStorage) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
