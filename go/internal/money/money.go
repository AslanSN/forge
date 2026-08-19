// Package money keeps amounts exact.
//
// The lesson of this package is a Go-specific one: the standard library has no
// decimal type. C# has `decimal`, so paso-01 there was only about *using* it.
// In Go you must choose a representation, and the choice is part of the step:
//
//	minor units (int64 cents)  → exact, no dependency, overflow is the only risk
//	a decimal library          → exact, richer, one more dependency to justify
//
// The scaffolding assumes minor units. That assumption is a decision, not a
// fact: if you decide otherwise, change the type first and let the tests follow,
// and write down *why* in docs/paso-02-*.md. What is NOT negotiable is that
// float64 never touches money — see TestFloatDrifts.
//
// Postgres side: balances stay numeric(18,2). The boundary is crossed as TEXT
// (`balance::text` in the SQL), never as a driver-chosen float. That is why
// ParseMinor and String exist.
//
// AUTHORSHIP (COLOPHON.md): ParseMinor/String/Scale are paso-00-level
// scaffolding — the numeric↔Go boundary, written for you so the step stays
// about concurrency. NormalizeAmount is paso-01 and is YOURS.
package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Scale is the number of decimal places Postgres stores: numeric(18,2).
const Scale = 2

// Minor is an exact monetary amount in minor units (cents).
// There is deliberately no constructor taking a float64.
type Minor int64

// String renders the amount the way Postgres would: "1234" → "12.34".
func (m Minor) String() string {
	neg := m < 0
	if neg {
		m = -m
	}
	units := int64(m) / 100
	cents := int64(m) % 100
	s := fmt.Sprintf("%d.%02d", units, cents)
	if neg {
		return "-" + s
	}
	return s
}

// ErrBadAmount is returned for anything that is not an exact 2-decimal amount.
var ErrBadAmount = errors.New("not an exact amount with at most 2 decimals")

// ParseMinor converts a decimal string to Minor, exactly.
// "12.34" → 1234, "12.3" → 1230, "12" → 1200. It rejects exponents, spaces,
// thousands separators and anything with more than two decimal places.
func ParseMinor(s string) (Minor, error) {
	// No trimming on purpose: whitespace means the caller passed something it
	// did not parse itself. Postgres never sends it; a form field might.
	if s == "" {
		return 0, ErrBadAmount
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")

	intPart, fracPart, hasFrac := strings.Cut(s, ".")
	if intPart == "" || (hasFrac && fracPart == "") {
		return 0, ErrBadAmount
	}
	if len(fracPart) > Scale {
		return 0, ErrBadAmount
	}
	for _, part := range []string{intPart, fracPart} {
		for _, r := range part {
			if r < '0' || r > '9' {
				return 0, ErrBadAmount
			}
		}
	}
	units, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, ErrBadAmount
	}
	cents := int64(0)
	if hasFrac {
		padded := fracPart + strings.Repeat("0", Scale-len(fracPart))
		cents, err = strconv.ParseInt(padded, 10, 64)
		if err != nil {
			return 0, ErrBadAmount
		}
	}
	total := units*100 + cents
	if neg {
		total = -total
	}
	return Minor(total), nil
}

// ── paso-01 · YOUR TURN ──────────────────────────────────────────────────────

// NormalizeAmount validates a monetary amount arriving from the outside world
// and returns it as Minor.
//
// Rules (make money_test.go green):
//   - the amount must parse exactly (reuse ParseMinor)
//   - it must be strictly greater than zero
//   - more than two decimal places is an error, never a rounding
//
// Return a non-nil error describing the problem; never return a partially
// valid amount alongside an error.
func NormalizeAmount(raw string) (Minor, error) {
	panic("paso-01: implement money.NormalizeAmount (delete this panic)")
}
