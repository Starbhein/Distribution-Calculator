package export

import (
	"os"
	"path/filepath"
	"testing"
)

// Approval tests pinning the CURRENT export title strings byte-for-byte
// (PR4c refactor guard, spec §3: the plot.go:224 formatParams switch is
// replaced by Spec.FormatParams with identical output). Green against the
// pre-refactor switch; MUST stay green — unmodified — after the refactor.
// Params mirror the real call shape: distribution params + trailing x.
func TestBuildTitleApproval(t *testing.T) {
	data := []float64{1, 2, 3}
	cases := []struct {
		name   string
		dist   string
		params []float64
		want   string
	}{
		{"binomial", "Binomial", []float64{0.5, 10, 3}, "Binomial\nn=3, μ̂=2.0000, σ̂=1.0000, p=0.5000, n=10"},
		{"poisson", "Poisson", []float64{4, 2}, "Poisson\nn=3, μ̂=2.0000, σ̂=1.0000, λ=4.0000"},
		{"hypergeometric", "Hypergeométrica", []float64{12, 3, 4, 1}, "Hypergeométrica\nn=3, μ̂=2.0000, σ̂=1.0000, N=12, M=3, n=4"},
		{"normal", "Normal", []float64{10, 2, 10}, "Normal\nn=3, μ̂=2.0000, σ̂=1.0000, μ=10.0000, σ=2.0000"},
		{"exponential lambda", "Exponencial (λ)", []float64{2, 1}, "Exponencial (λ)\nn=3, μ̂=2.0000, σ̂=1.0000, λ=2.0000"},
		{"exponential beta", "Exponencial (β)", []float64{2, 1}, "Exponencial (β)\nn=3, μ̂=2.0000, σ̂=1.0000, β=2.0000"},
		{"bernoulli", "Bernoulli", []float64{0.3, 1}, "Bernoulli\nn=3, μ̂=2.0000, σ̂=1.0000, p=0.3000"},
		{"geometric", "Geométrica", []float64{0.25, 2}, "Geométrica\nn=3, μ̂=2.0000, σ̂=1.0000, p=0.2500"},
		{"uniform", "Uniforme continua", []float64{2, 4, 3}, "Uniforme continua\nn=3, μ̂=2.0000, σ̂=1.0000, a=2.0000, b=4.0000"},
		{"unknown name formats empty", "Desconocida", []float64{1}, "Desconocida\nn=3, μ̂=2.0000, σ̂=1.0000, "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildTitle(tc.dist, tc.params, data); got != tc.want {
				t.Errorf("buildTitle changed.\n got: %q\nwant: %q", got, tc.want)
			}
		})
	}
}

// TestExportPlotRuntimeHarness exercises the real export boundary end to
// end (runtime harness for the PR4c refactor): a discrete and a continuous
// export each produce a non-empty PNG and SVG via t.TempDir. Values are
// pinned separately (registry funcs tests + buildTitle approval); this test
// proves the dispatch rewiring keeps the render pipeline working.
func TestExportPlotRuntimeHarness(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name   string
		data   []float64
		dist   string
		params []float64
		markX  float64
	}{
		{"discrete binomial", []float64{2, 3, 3, 4, 5, 3, 4, 2, 3, 4}, "Binomial", []float64{0.5, 10, 3}, 3},
		{"discrete hypergeometric", []float64{0, 1, 1, 0, 2, 1, 0, 1, 1, 0}, "Hypergeométrica", []float64{12, 3, 4, 1}, 1},
		{"continuous normal", []float64{8, 9, 10, 11, 12, 9.5, 10.5, 10, 9, 11}, "Normal", []float64{10, 2, 10}, 10},
		{"continuous uniform", []float64{2.2, 2.8, 3.1, 3.9, 2.5, 3.3, 2.9, 3.6, 2.4, 3.8}, "Uniforme continua", []float64{2, 4, 3}, 3},
	}
	for _, tc := range cases {
		for _, format := range []string{"png", "svg"} {
			t.Run(tc.name+"/"+format, func(t *testing.T) {
				out := filepath.Join(dir, "plot."+format)
				if err := ExportPlot(tc.data, tc.dist, tc.params, tc.markX, out, format); err != nil {
					t.Fatalf("ExportPlot returned error: %v", err)
				}
				info, err := os.Stat(out)
				if err != nil {
					t.Fatalf("output file missing: %v", err)
				}
				if info.Size() == 0 {
					t.Error("output file is empty")
				}
			})
		}
	}
}
