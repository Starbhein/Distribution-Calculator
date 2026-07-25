package ui

import "testing"

// Approval tests pinning the CURRENT rendered chart output byte-for-byte
// (PR4c refactor guard, spec §5 "displayed values are unchanged"). They are
// green against the pre-refactor name-keyed switches and MUST stay green —
// unmodified — after the switches are replaced by registry dispatch
// (ByName + PMFFunc/PDFFunc). The Poisson case deliberately includes k=5
// with lambda=0.5 so the out-of-row tail fallback ("teo:0.0002") is pinned.
func TestRenderChartApproval(t *testing.T) {
	cases := []struct {
		name string
		got  func() string
		want string
	}{
		{
			name: "binomial discrete with theory (UI order (p, n), trailing x)",
			got: func() string {
				return RenderDiscreteHistogram([]float64{2, 3, 3, 4, 5, 3, 4, 2, 3, 4}, "Binomial", []float64{0.5, 10, 3}, 3, 80, 0)
			},
			want: "   2 |██████████████████████████ teo:0.0439 emp:0.2000\n   3 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ teo:0.1172 emp:0.4000 ← tu x\n   4 |███████████████████████████████████████ teo:0.2051 emp:0.3000\n   5 |█████████████ teo:0.2461 emp:0.1000\n",
		},
		{
			name: "hypergeometric discrete with theory (asymmetric N=12,M=3,n=4)",
			got: func() string {
				return RenderDiscreteHistogram([]float64{0, 1, 1, 0, 2, 1, 0, 1, 1, 0}, "Hypergeométrica", []float64{12, 3, 4, 1}, 1, 80, 0)
			},
			want: "   0 |█████████████████████████████████████████ teo:0.2545 emp:0.4000\n   1 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ teo:0.5091 emp:0.5000 ← tu x\n   2 |██████████ teo:0.2182 emp:0.1000\n",
		},
		{
			name: "poisson discrete with out-of-row tail k=5",
			got: func() string {
				return RenderDiscreteHistogram([]float64{0, 1, 0, 2, 1, 0, 1, 3, 1, 5}, "Poisson", []float64{0.5, 1}, 1, 80, 0)
			},
			want: "   0 |███████████████████████████████████████ teo:0.6065 emp:0.3000\n   1 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ teo:0.3033 emp:0.4000 ← tu x\n   2 |█████████████ teo:0.0758 emp:0.1000\n   3 |█████████████ teo:0.0126 emp:0.1000\n   5 |█████████████ teo:0.0002 emp:0.1000\n",
		},
		{
			name: "geometric discrete with theory (closed-form path)",
			got: func() string {
				return RenderDiscreteHistogram([]float64{1, 2, 1, 3, 1, 2, 4, 1, 2, 1}, "Geométrica", []float64{0.25, 2}, 2, 80, 0)
			},
			want: "   1 |████████████████████████████████████████████████████ teo:0.2500 emp:0.5000\n   2 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ teo:0.1875 emp:0.3000 ← tu x\n   3 |██████████ teo:0.1406 emp:0.1000\n   4 |██████████ teo:0.1055 emp:0.1000\n",
		},
		{
			name: "normal continuous with theory",
			got: func() string {
				return RenderContinuousHistogram([]float64{8, 9, 10, 11, 12, 9.5, 10.5, 10, 9, 11}, "Normal", []float64{10, 2, 10}, 10, 80, 0)
			},
			want: "     8.0-8.4 |█████████████████████ dens:0.250 pdf:0.133\n     8.4-8.8 | dens:0.000 pdf:0.156\n     8.8-9.2 |███████████████████████████████████████████ dens:0.500 pdf:0.176\n     9.2-9.6 |█████████████████████ dens:0.250 pdf:0.191\n    9.6-10.0 | dens:0.000 pdf:0.198\n   10.0-10.4 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ dens:0.500 pdf:0.198 ← tu x\n   10.4-10.8 |█████████████████████ dens:0.250 pdf:0.191\n   10.8-11.2 |███████████████████████████████████████████ dens:0.500 pdf:0.176\n   11.2-11.6 | dens:0.000 pdf:0.156\n   11.6-12.0 |█████████████████████ dens:0.250 pdf:0.133\n",
		},
		{
			name: "uniform continuous with theory",
			got: func() string {
				return RenderContinuousHistogram([]float64{2.2, 2.8, 3.1, 3.9, 2.5, 3.3, 2.9, 3.6, 2.4, 3.8}, "Uniforme continua", []float64{2, 4, 3}, 3, 80, 0)
			},
			want: "     2.2-2.4 |█████████████████████ dens:0.588 pdf:0.500\n     2.4-2.5 |███████████████████████████████████████████ dens:1.176 pdf:0.500\n     2.5-2.7 | dens:0.000 pdf:0.500\n     2.7-2.9 |█████████████████████ dens:0.588 pdf:0.500\n     2.9-3.0 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ dens:0.588 pdf:0.500 ← tu x\n     3.0-3.2 |█████████████████████ dens:0.588 pdf:0.500\n     3.2-3.4 |█████████████████████ dens:0.588 pdf:0.500\n     3.4-3.6 | dens:0.000 pdf:0.500\n     3.6-3.7 |█████████████████████ dens:0.588 pdf:0.500\n     3.7-3.9 |███████████████████████████████████████████ dens:1.176 pdf:0.500\n",
		},
		{
			name: "discrete without theory",
			got:  func() string { return RenderHistogram([]float64{1, 2, 2, 3}, true, 80, 2) },
			want: "   1 |██████████████████████████ 25.0%\n   2 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 50.0% ← tu x\n   3 |██████████████████████████ 25.0%\n",
		},
		{
			name: "continuous without theory",
			got:  func() string { return RenderHistogram([]float64{1, 2, 3, 4, 5}, false, 80, 3) },
			want: "     1.0-1.8 |███████████████████████████████████████████ 20.0%\n     1.8-2.6 |███████████████████████████████████████████ 20.0%\n     2.6-3.4 |▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓ 20.0% ← tu x\n     3.4-4.2 |███████████████████████████████████████████ 20.0%\n     4.2-5.0 |███████████████████████████████████████████ 20.0%\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.got(); got != tc.want {
				t.Errorf("rendered output changed.\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}
