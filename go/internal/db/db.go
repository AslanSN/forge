// Package db resolves the Postgres connection and hands out a pool.
//
// paso-00-level scaffolding (written for you): the same binary must talk to the
// docker-compose database locally and to a real one elsewhere, so the DSN comes
// from the environment and is never hard-coded.
//
// The .NET side reads FORGE_DB in ADO.NET format ("Host=…;Port=…"). pgx wants a
// URL, so the Go line reads FORGE_DB_URL and falls back to the compose defaults,
// honouring FORGE_DB_PORT from .env the way docker-compose.yml does.
package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DSN returns the Postgres URL for this machine.
func DSN() string {
	if url := os.Getenv("FORGE_DB_URL"); url != "" {
		return url
	}
	port := os.Getenv("FORGE_DB_PORT")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf("postgres://forge:forge@localhost:%s/forge?sslmode=disable", port)
}

// Pool opens a connection pool. maxConns is explicit and small on purpose:
// paso-02 asks you to notice what happens when more goroutines want a
// connection than the pool has.
func Pool(ctx context.Context, maxConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(DSN())
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = maxConns
	return pgxpool.NewWithConfig(ctx, cfg)
}

// TestPool gives a test a pool, or skips the test when the database is down —
// the same contract as the .NET integration tests ("skip if the db is down"),
// so `go test ./...` stays useful without Docker running.
func TestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := Pool(ctx, 10)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err != nil {
		if pool != nil {
			pool.Close()
		}
		t.Skipf("no database at %s (run `make up && make migrate`): %v", DSN(), err)
	}
	t.Cleanup(pool.Close)
	return pool
}
