package accounts_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/AslanSN/forge/go/internal/money"
)

// paso-02 · the executable spec.
//
// These tests are the failure you are here to feel. They are written against a
// contract that is not wrong — "a ledger never loses an update and never goes
// negative" — and the naive read-check-write implementation from paso-01 breaks
// them the moment more than one goroutine is involved.
//
// Read docs/paso-02-transactions-isolation-and-the-double-spend-race.md before
// changing a line of the implementation. Run them with the race detector too
// (`make go-race`) and notice what it does NOT tell you.

// TestConcurrentWithdrawals_NeverDoubleSpend is the double spend itself.
// One account holds 100.00 and 20 goroutines each try to take 60.00.
// Arithmetic, not opinion: exactly one can succeed.
func TestConcurrentWithdrawals_NeverDoubleSpend(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	id := seedAccount(t, pool, 10000) // 100.00
	skipUntilImplemented(t, func() { _, _ = store.Withdraw(ctx, id, 1) })

	const goroutines = 20
	var (
		wg        sync.WaitGroup
		succeeded atomic.Int64
	)
	start := make(chan struct{}) // release them all at once, to widen the window
	for range goroutines {
		wg.Go(func() {
			<-start
			if _, err := store.Withdraw(ctx, id, 6000); err == nil { // 60.00
				succeeded.Add(1)
			}
		})
	}
	close(start)
	wg.Wait()

	final := balanceOf(t, store, id)
	if n := succeeded.Load(); n != 1 {
		t.Errorf("%d withdrawals of 60.00 succeeded against a balance of 100.00; want exactly 1", n)
	}
	if final != 4000 {
		t.Errorf("final balance = %s; want 40.00", final)
	}
	if final < 0 {
		t.Errorf("the ledger went negative (%s) — money was created out of nothing", final)
	}
}

// TestConcurrentDeposits_LoseNoUpdates is the lost update, the same race wearing
// different clothes: nothing is ever refused here, and money still disappears.
func TestConcurrentDeposits_LoseNoUpdates(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	id := seedAccount(t, pool, 0)
	skipUntilImplemented(t, func() { _, _ = store.Deposit(ctx, id, 1) })

	const (
		goroutines = 50
		each       = money.Minor(100) // 1.00
	)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range goroutines {
		wg.Go(func() {
			<-start
			if _, err := store.Deposit(ctx, id, each); err != nil {
				t.Errorf("deposit: %v", err)
			}
		})
	}
	close(start)
	wg.Wait()

	want := money.Minor(goroutines) * each
	if got := balanceOf(t, store, id); got != want {
		t.Errorf("balance = %s; want %s — %d deposits went in, some were overwritten", got, want, goroutines)
	}
}

// TestConcurrentTransfers_ConserveTotal is paso-02's second half. Money moves in
// both directions at once: the total across both accounts is an invariant, and
// two goroutines locking the same two rows in opposite order is how you meet a
// deadlock (SQLSTATE 40P01) for the first time.
func TestConcurrentTransfers_ConserveTotal(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	a := seedAccount(t, pool, 10000) // 100.00
	b := seedAccount(t, pool, 10000) // 100.00
	skipUntilImplemented(t, func() { _ = store.Transfer(ctx, a, b, 1) })

	const rounds = 25
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range rounds {
		wg.Go(func() {
			<-start
			if err := store.Transfer(ctx, a, b, 500); err != nil { // 5.00
				t.Errorf("transfer a→b: %v", err)
			}
		})
		wg.Go(func() {
			<-start
			if err := store.Transfer(ctx, b, a, 500); err != nil { // 5.00
				t.Errorf("transfer b→a: %v", err)
			}
		})
	}
	close(start)
	wg.Wait()

	total := balanceOf(t, store, a) + balanceOf(t, store, b)
	if total != 20000 {
		t.Errorf("total across both accounts = %s; want 200.00 — a transfer was not atomic", total)
	}
}

// TestTransferIsAtomic checks the half nobody tests: a transfer that must fail
// leaves BOTH sides untouched, not just the one that errored.
func TestTransferIsAtomic(t *testing.T) {
	store, pool := newStore(t)
	ctx := context.Background()
	a := seedAccount(t, pool, 1000)  // 10.00
	b := seedAccount(t, pool, 10000) // 100.00
	skipUntilImplemented(t, func() { _ = store.Transfer(ctx, a, b, 1) })

	if err := store.Transfer(ctx, a, b, 5000); err == nil { // 50.00 out of 10.00
		t.Fatal("transferring more than the source holds must fail")
	}
	if got := balanceOf(t, store, a); got != 1000 {
		t.Errorf("source moved on a failed transfer: %s; want 10.00", got)
	}
	if got := balanceOf(t, store, b); got != 10000 {
		t.Errorf("destination moved on a failed transfer: %s; want 100.00", got)
	}
}
