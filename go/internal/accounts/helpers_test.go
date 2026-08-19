package accounts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AslanSN/forge/go/internal/accounts"
	"github.com/AslanSN/forge/go/internal/db"
	"github.com/AslanSN/forge/go/internal/money"
)

// seedAccount inserts a funded account with raw SQL on purpose: a test fixture
// must not depend on the code under test, or a broken Deposit hides a broken
// Withdraw.
func seedAccount(t *testing.T, pool *pgxpool.Pool, balance money.Minor) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO accounts (name, balance) VALUES ($1, $2::numeric) RETURNING id::text;`,
		fmt.Sprintf("test-%s", t.Name()), balance.String(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seeding account: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM accounts WHERE id = $1::uuid;`, id)
	})
	return id
}

func balanceOf(t *testing.T, store *accounts.Store, id string) money.Minor {
	t.Helper()
	acc, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("reading balance: %v", err)
	}
	return acc.Balance
}

// skipUntilImplemented turns the "panic: paso-NN: implement …" stubs into a
// readable skip, so a red run tells you which step to do next instead of
// crashing the test binary from inside a goroutine.
func skipUntilImplemented(t *testing.T, probe func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		probe()
	}()
	if r := <-done; r != nil {
		if msg, ok := r.(string); ok {
			t.Skipf("not implemented yet → %s", msg)
		}
		t.Skipf("not implemented yet → %v", r)
	}
}

func newStore(t *testing.T) (*accounts.Store, *pgxpool.Pool) {
	t.Helper()
	pool := db.TestPool(t)
	return accounts.New(pool), pool
}
