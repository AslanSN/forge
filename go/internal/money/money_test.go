package money_test

import (
	"testing"

	"github.com/AslanSN/forge/go/internal/money"
)

// TestFloatDrifts is the paso-01 demonstration, and it PASSES: it is here to be
// read, not fixed. Adding 0.10 ten times in float64 is not 1.00. The same sum in
// minor units is exact. This is why money is numeric in Postgres and int64 here.
func TestFloatDrifts(t *testing.T) {
	var f float64
	for range 10 {
		f += 0.1
	}
	if f == 1.0 {
		t.Fatalf("expected float64 to drift, got exactly %v — read the test again", f)
	}
	t.Logf("float64 sum of ten 0.1 = %.20f", f)

	var exact money.Minor
	for range 10 {
		exact += 10 // 0.10 in cents
	}
	if exact != 100 {
		t.Fatalf("minor units must be exact: got %d, want 100", exact)
	}
}

func TestParseMinor(t *testing.T) {
	ok := map[string]money.Minor{
		"12.34": 1234, "12.3": 1230, "12": 1200, "0.01": 1, "0": 0, "-5.05": -505,
	}
	for in, want := range ok {
		got, err := money.ParseMinor(in)
		if err != nil || got != want {
			t.Errorf("ParseMinor(%q) = %d, %v; want %d, nil", in, got, err, want)
		}
	}
	bad := []string{"", "abc", "1.234", "1e3", "1,00", " ", "1.", ".5", "10 "}
	for _, in := range bad {
		if got, err := money.ParseMinor(in); err == nil {
			t.Errorf("ParseMinor(%q) = %d, nil; want an error", in, got)
		}
	}
}

func TestMinorString(t *testing.T) {
	cases := map[money.Minor]string{1234: "12.34", 5: "0.05", 0: "0.00", -505: "-5.05"}
	for in, want := range cases {
		if got := in.String(); got != want {
			t.Errorf("Minor(%d).String() = %q; want %q", int64(in), got, want)
		}
	}
}

// ── paso-01 · RED until you implement money.NormalizeAmount ───────────────────

func TestNormalizeAmount_accepts(t *testing.T) {
	for _, in := range []string{"0.01", "10", "10.5", "999999.99"} {
		got, err := money.NormalizeAmount(in)
		if err != nil {
			t.Errorf("NormalizeAmount(%q) errored: %v", in, err)
			continue
		}
		want, _ := money.ParseMinor(in)
		if got != want {
			t.Errorf("NormalizeAmount(%q) = %d; want %d", in, got, want)
		}
	}
}

func TestNormalizeAmount_rejects(t *testing.T) {
	// Zero and negatives are not deposits or withdrawals; three decimals is a
	// caller bug, not something to round away silently.
	for _, in := range []string{"0", "0.00", "-1", "-0.01", "1.234", "abc", ""} {
		if got, err := money.NormalizeAmount(in); err == nil {
			t.Errorf("NormalizeAmount(%q) = %d, nil; want an error", in, got)
		}
	}
}
