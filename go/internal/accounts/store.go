// Package accounts is the ledger's write path.
//
// No ORM: raw SQL over pgx, always parameterized ($1, $2 …), never string
// concatenation. Values cross the boundary as TEXT (`id::text`, `balance::text`)
// so the driver never decides for you that money is a float.
//
// AUTHORSHIP (COLOPHON.md): Create/Get/NormalizeName are the paso-00 reference,
// ported from the C# line so the Go line starts where that one did. Everything
// under "YOUR TURN" is yours to type by hand.
//
// NOTE, and it is the whole plot: `accounts.balance` is a MUTABLE column. That
// is deliberately naive. paso-02 runs concurrent withdrawals against it and
// makes it leak money; the fix ends in an append-only `entries` table where the
// balance is derived rather than overwritten.
package accounts

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AslanSN/forge/go/internal/money"
)

const MaxNameLength = 100

var (
	ErrNotFound          = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidName       = errors.New("name is required and must be at most 100 characters")
)

// Account is what the store returns. Balance never leaves as a float.
type Account struct {
	ID      string
	Name    string
	Balance money.Minor
}

// Store owns the pool, not a connection: every call borrows one.
type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// NormalizeName is pure logic — no I/O — so it is unit-testable without a DB.
func NormalizeName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len(name) > MaxNameLength {
		return "", ErrInvalidName
	}
	return name, nil
}

func (s *Store) Create(ctx context.Context, rawName string) (Account, error) {
	name, err := NormalizeName(rawName)
	if err != nil {
		return Account{}, err
	}
	const q = `
		INSERT INTO accounts (name)
		VALUES ($1)
		RETURNING id::text, name, balance::text;`
	return scanAccount(s.pool.QueryRow(ctx, q, name))
}

func (s *Store) Get(ctx context.Context, id string) (Account, error) {
	const q = `
		SELECT id::text, name, balance::text
		FROM accounts
		WHERE id = $1::uuid;`
	acc, err := scanAccount(s.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	return acc, err
}

// scanAccount is the numeric↔Go boundary in one place: text in, exact out.
func scanAccount(row pgx.Row) (Account, error) {
	var (
		acc     Account
		balance string
	)
	if err := row.Scan(&acc.ID, &acc.Name, &balance); err != nil {
		return Account{}, err
	}
	amount, err := money.ParseMinor(balance)
	if err != nil {
		return Account{}, err
	}
	acc.Balance = amount
	return acc, nil
}

// ── paso-01 · YOUR TURN · money, the naive way ───────────────────────────────
//
// Implement Deposit and Withdraw so store_test.go goes green. Validate the
// amount with money.NormalizeAmount. Withdraw must refuse to overdraw and
// return ErrInsufficientFunds.
//
// Do it the TEMPTING way on purpose: read the balance, check it here in Go,
// then write the new balance back. It passes every single-threaded test in
// store_test.go — and race_test.go exists to prove that is not enough.
// Leaving the bug in deliberately is the method (see AGENTS.md).

func (s *Store) Deposit(ctx context.Context, id string, amount money.Minor) (Account, error) {
	if amount <= 0 {
		return Account{}, money.ErrBadAmount
	}

	acc, err := s.Get(ctx, id)
	if err != nil {
		return Account{}, err
	}

	acc.Balance = acc.Balance + amount

	q := `
	    UPDATE accounts
		SET balance = $1::numeric
		WHERE id = $2::uuid
		RETURNING id::text, name, balance::text;
	`

	updAcc, err := scanAccount(s.pool.QueryRow(ctx, q, acc.Balance.String(), id))

	if err != nil {
		return Account{}, err
	}

	return updAcc, err

}

func (s *Store) Withdraw(ctx context.Context, id string, amount money.Minor) (Account, error) {
	if amount <= 0 {
		return Account{}, money.ErrBadAmount
	}

	acc, err := s.Get(ctx, id)
	if err != nil {
		return Account{}, err
	}

	if amount > acc.Balance {
		return Account{}, ErrInsufficientFunds
	}

	acc.Balance = acc.Balance - amount

	q := `
	    UPDATE accounts
		SET balance = $1::numeric
		WHERE id = $2::uuid
		RETURNING id::text, name, balance::text
	`

	updAcc, err := scanAccount(s.pool.QueryRow(ctx, q, acc.Balance.String(), id))

	if err != nil {
		return Account{}, err
	}

	return updAcc, err
}

// ── paso-02 · YOUR TURN · the double-spend race ──────────────────────────────
//
// Transfer moves money between two accounts. It must be atomic (both legs or
// neither) and must never let an account go negative, including when many
// goroutines transfer at once and when two of them transfer in opposite
// directions at the same time.
//
// Do NOT start here. Start by running race_test.go against your naive Withdraw,
// watching it fail, and reading docs/paso-02-*.md. Transfer is where the fix you
// choose there gets written down.

func (s *Store) Transfer(ctx context.Context, fromID, toID string, amount money.Minor) error {
	panic("paso-02: implement Transfer (see docs/paso-02-transactions-isolation-and-the-double-spend-race.md)")
}
