package distributions

import "testing"

// Backstop tests for the cross-parameter rules M<=N and n<=N on
// NewHypergeometric (design §7, spec §3 validation convergence). These are
// defense for direct library consumers; the registry Validate layer is THE
// app-level rule. Lives in its own file so the pinned
// hypergeometric_test.go stays unmodified (spec §8).
func TestNewHypergeometricCrossParamBackstop(t *testing.T) {
	cases := []struct {
		name      string
		successes int // M
		pop       int // N
		sample    int // n
	}{
		{"M greater than N", 13, 12, 4},
		{"n greater than N", 3, 12, 13},
		{"both greater than N", 13, 12, 13},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewHypergeometric(tc.successes, tc.pop, tc.sample)
			if err == nil {
				t.Fatalf("NewHypergeometric(M=%d, N=%d, n=%d) expected error, got nil (struct %+v)",
					tc.successes, tc.pop, tc.sample, h)
			}
			if h != nil {
				t.Errorf("expected nil struct on error, got %+v", h)
			}
		})
	}
}

func TestNewHypergeometricBackstopAcceptsValidCombos(t *testing.T) {
	// Regression: parameter sets accepted today remain accepted (spec §3).
	cases := []struct {
		name      string
		successes int
		pop       int
		sample    int
	}{
		{"pinned case", 3, 12, 4},
		{"M equals N", 12, 12, 4},
		{"n equals N", 3, 12, 12},
		{"all equal", 5, 5, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewHypergeometric(tc.successes, tc.pop, tc.sample)
			if err != nil {
				t.Fatalf("NewHypergeometric(M=%d, N=%d, n=%d) returned error: %v",
					tc.successes, tc.pop, tc.sample, err)
			}
			if h == nil {
				t.Fatal("expected non-nil struct")
			}
		})
	}
}
