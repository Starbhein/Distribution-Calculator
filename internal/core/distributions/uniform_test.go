package distributions

import (
	"math"
	"testing"
)

// Tests for the minimal Uniform struct (design §1.3b carve-out, proposal §4).
// Formulas are pinned to the values the UI computes inline today at
// internal/ui/theoretical.go:122-137,321-340 so displayed values are unchanged.

func TestUniformMoments(t *testing.T) {
	t.Run("U(2,4) avg variance stddev", func(t *testing.T) {
		u, err := NewUniform(2, 4)
		if err != nil {
			t.Fatalf("NewUniform(2, 4) returned error: %v", err)
		}
		evalFloats(t, u.Avg(), 3.0)
		evalFloats(t, u.Variance(), 4.0/12.0)
		evalFloats(t, u.StdDev(), math.Sqrt(4.0/12.0))
	})

	t.Run("U(0,1) avg variance stddev", func(t *testing.T) {
		u, err := NewUniform(0, 1)
		if err != nil {
			t.Fatalf("NewUniform(0, 1) returned error: %v", err)
		}
		evalFloats(t, u.Avg(), 0.5)
		evalFloats(t, u.Variance(), 1.0/12.0)
		evalFloats(t, u.StdDev(), math.Sqrt(1.0/12.0))
	})
}

func TestUniformPDF(t *testing.T) {
	u, err := NewUniform(2, 4)
	if err != nil {
		t.Fatalf("NewUniform(2, 4) returned error: %v", err)
	}

	cases := []struct {
		name string
		x    float64
		want float64
	}{
		{"below a", 1.0, 0.0},
		{"at a", 2.0, 0.5},
		{"inside", 3.0, 0.5},
		{"at b", 4.0, 0.5},
		{"above b", 5.0, 0.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := u.PDF(tc.x)
			if err != nil {
				t.Fatalf("PDF(%v) returned error: %v", tc.x, err)
			}
			evalFloats(t, got, tc.want)
		})
	}
}

func TestUniformCDF(t *testing.T) {
	u, err := NewUniform(2, 4)
	if err != nil {
		t.Fatalf("NewUniform(2, 4) returned error: %v", err)
	}

	cases := []struct {
		name string
		x    float64
		want float64
	}{
		{"below a", 1.0, 0.0},
		{"at a", 2.0, 0.0},
		{"inside", 3.0, 0.5},
		{"at b", 4.0, 1.0},
		{"above b", 5.0, 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := u.CDF(tc.x)
			if err != nil {
				t.Fatalf("CDF(%v) returned error: %v", tc.x, err)
			}
			evalFloats(t, got, tc.want)
		})
	}
}

func TestUniformInvalidParams(t *testing.T) {
	t.Run("a greater than b", func(t *testing.T) {
		u, err := NewUniform(4, 2)
		if err == nil {
			t.Fatal("NewUniform(4, 2) expected error, got nil")
		}
		if u != nil {
			t.Fatalf("NewUniform(4, 2) expected nil struct, got %+v", u)
		}
	})

	t.Run("a equal to b", func(t *testing.T) {
		u, err := NewUniform(3, 3)
		if err == nil {
			t.Fatal("NewUniform(3, 3) expected error, got nil")
		}
		if u != nil {
			t.Fatalf("NewUniform(3, 3) expected nil struct, got %+v", u)
		}
	})
}
