package distributions

import (
	"errors"
	"math"
)

// Uniform represents the continuous uniform distribution on [A, B].
// Introduced so the registry can dispatch "Uniforme continua" through the
// same struct-based path as every other distribution (design §1.3b
// carve-out, proposal §4). The formulas are copied exactly from the inline
// UI copies at internal/ui/theoretical.go:122-137,321-340 so displayed
// values are unchanged.
type Uniform struct {
	A float64
	B float64
}

func NewUniform(a, b float64) (*Uniform, error) {
	if a >= b {
		return nil, errors.New("a must be lower than b")
	}
	return &Uniform{A: a, B: b}, nil
}

func (u Uniform) Avg() float64 {
	return (u.A + u.B) / 2.0
}

func (u Uniform) Variance() float64 {
	width := u.B - u.A
	return (width * width) / 12.0
}

func (u Uniform) StdDev() float64 {
	return math.Sqrt(u.Variance())
}

func (u Uniform) PDF(x float64) (float64, error) {
	width := u.B - u.A
	if x < u.A {
		return 0, nil
	}
	if x > u.B {
		return 0, nil
	}
	return 1.0 / width, nil
}

func (u Uniform) CDF(x float64) (float64, error) {
	width := u.B - u.A
	if x < u.A {
		return 0, nil
	}
	if x > u.B {
		return 1, nil
	}
	return (x - u.A) / width, nil
}

// Compile-time conformance assertion (design §1.3b).
var _ ContinuousDistribution = (*Uniform)(nil)
