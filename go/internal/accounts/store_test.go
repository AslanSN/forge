package accounts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AslanSN/forge/go/internal/accounts"
	"github.com/AslanSN/forge/go/internal/money"
)

// ── paso-00 reference · these pass already ───────────────────────────────────

func TestNormalizeName(t *testing.T) {
	if _, err := accounts.NormalizeName("  Alan  "); err != nil {
		t.Errorf("trimmed name should be valid: %v", err)
	}
	for _, bad := range []string{"", "   ", string(make([]byte, 101))} {
		if _, err := accounts.NormalizeName(bad); err == nil {
			t.Errorf("NormalizeName(%q) should have failed", bad)
		}
	}
}

func TestCreateAndGet(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()

	acc, err := store.Create(ctx, "reference account")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM accounts WHERE id = $1::uuid;`, acc.ID)
	})

	got, err := store.Get(ctx, acc.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != acc.ID || got.Name != "reference account" || got.Balance != 0 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if _, err := store.Get(ctx, nonexistentID); !errors.Is(err, accounts.ErrNotFound) {
		t.Errorf("missing account should be ErrNotFound, got %v", err)
	}
}

// ── paso-01 · RED until Deposit/Withdraw exist ───────────────────────────────

func TestDepositAddsExactly(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	id := seedAccount(t, pool, 0)
	requireImplemented(t, func() { _, _ = store.Deposit(ctx, nonexistentID, 1) })

	for range 3 {
		if _, err := store.Deposit(ctx, id, 1050); err != nil { // 10.50
			t.Fatalf("deposit: %v", err)
		}
	}
	if got := balanceOf(t, store, id); got != 3150 {
		t.Errorf("balance = %s; want 31.50 (three deposits of 10.50)", got)
	}
}

func TestWithdrawRefusesToOverdraw(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	id := seedAccount(t, pool, 10000) // 100.00
	requireImplemented(t, func() { _, _ = store.Withdraw(ctx, nonexistentID, 1) })

	if _, err := store.Withdraw(ctx, id, 10001); !errors.Is(err, accounts.ErrInsufficientFunds) {
		t.Errorf("overdraw should be ErrInsufficientFunds, got %v", err)
	}
	if got := balanceOf(t, store, id); got != 10000 {
		t.Errorf("a refused withdrawal must not move money: balance = %s", got)
	}
	if _, err := store.Withdraw(ctx, id, 6000); err != nil {
		t.Fatalf("valid withdrawal: %v", err)
	}
	if got := balanceOf(t, store, id); got != 4000 {
		t.Errorf("balance = %s; want 40.00", got)
	}
}

func TestAmountsAreValidated(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	id := seedAccount(t, pool, 10000)
	requireImplemented(t, func() { _, _ = store.Deposit(ctx, nonexistentID, 1) })

	for _, bad := range []money.Minor{0, -1} {
		if _, err := store.Deposit(ctx, id, bad); err == nil {
			t.Errorf("Deposit(%d) should have failed", bad)
		}
		if _, err := store.Withdraw(ctx, id, bad); err == nil {
			t.Errorf("Withdraw(%d) should have failed", bad)
		}
	}
}
