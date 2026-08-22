package accounts_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AslanSN/forge/go/internal/accounts"
	"github.com/AslanSN/forge/go/internal/db"
	"github.com/AslanSN/forge/go/internal/money"
)

// currentPaso is the step being worked on right now. It is the one line to bump
// when you move on, and it decides how an unimplemented stub reports:
//
//	stub from currentPaso   → FAIL. That is the exercise; the step starts red.
//	stub from a later step  → skip, so finishing this step still gets you green.
const currentPaso = "paso-01"

// nonexistentID is a well-formed UUID that no row has. Probing against it
// reaches the code under test without matching anything.
const nonexistentID = "00000000-0000-0000-0000-000000000000"

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

// stubPanic matches the "paso-NN: implement …" message the stubs panic with,
// capturing the step it belongs to.
var stubPanic = regexp.MustCompile(`^(paso-[0-9]{2}[a-z]?): `)

// requireImplemented runs probe to find out whether the stub it calls is still
// a panic, and reports accordingly (see currentPaso).
//
// The probe MUST be side-effect free — call it against nonexistentID, never
// against the account the test is about to measure. An earlier version probed
// with a real one-cent operation on the seeded account, and the moment the
// stubs were implemented every balance assertion in the suite came out a cent
// off, including race_test.go's, which then blamed the drift on a lost update.
// A test fixture that moves money cannot be used to test money.
func requireImplemented(t *testing.T, probe func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		probe()
	}()

	r := <-done
	if r == nil {
		return // implemented — carry on
	}
	msg := fmt.Sprint(r)
	step := stubPanic.FindStringSubmatch(msg)
	if step == nil {
		// Not a stub: something under test genuinely blew up. Never swallow it.
		t.Fatalf("probe panicked: %v", r)
	}
	if step[1] == currentPaso {
		t.Fatalf("not implemented yet → %s", msg)
	}
	t.Skipf("waiting on %s → %s", step[1], msg)
}

func newStore(t *testing.T) (*accounts.Store, *pgxpool.Pool) {
	t.Helper()
	pool := db.TestPool(t)
	return accounts.New(pool), pool
}
