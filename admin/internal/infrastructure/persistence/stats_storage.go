// Package persistence implements admin's StatsStorage port via pgxpool.
// All queries are READ-ONLY against the shared `hr` database — admin is
// the only service that crosses per-service table ownership; documented
// as a deliberate operational exception so adding aggregate counters
// doesn't require a new RPC in every domain service.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatsStorage holds a pgxpool connected to the shared `hr` database.
// Closed via Close() during graceful shutdown.
type StatsStorage struct {
	db *pgxpool.Pool
}

// NewStatsStorage initialises a pool, performs a Ping to fail fast if
// the DB is unreachable. Bounded by a 30-second context — same as other
// services.
func NewStatsStorage(connString string) (*StatsStorage, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}
	// Read-only workload, small pool is plenty.
	cfg.MaxConns = 4

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("init pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}
	return &StatsStorage{db: pool}, nil
}

func (s *StatsStorage) Close() {
	if s.db != nil {
		s.db.Close()
	}
}
