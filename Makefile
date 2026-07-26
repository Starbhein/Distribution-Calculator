# DistCalc developer loop. All targets are informational; there are no CI gates.
.PHONY: build run test test-short bench bench-precision bench-compare coverage

# Build the TUI binary.
build:
	go build ./cmd/tui

# Run the TUI.
run:
	go run ./cmd/tui

# Full test suite (canonical: go test ./...).
test:
	go test ./...

# Full test suite in short mode (skips testing.Short()-gated tests).
test-short:
	go test -short ./...

# All benchmarks with pinned flags for comparable runs (-benchtime=1s -count=5).
bench:
	go test -run='^$$' -bench=. -benchtime=1s -count=5 ./internal/core/distmath/ ./internal/core/distributions/ ./internal/core/distributions/sim/

# Precision evidence: pinned values, triangulation, and closed-form
# cross-checks, plus BOTH geometric moment tests. The -run regex hits the 1e5
# test and the 1e7 heavy variant; the heavy leg is gated behind
# DISTCALC_HEAVY, which this target sets inline so it always runs here.
bench-precision:
	go test -run 'Pinned|Triangulation|MatchesClosedForm' ./internal/core/distmath/
	DISTCALC_HEAVY=1 go test -run 'TestFillGeometricSmallPMoments' ./internal/core/distributions/sim/

# Optional stdlib-only Python algorithm comparison; skips cleanly if python3
# is not installed.
bench-compare:
	@command -v python3 >/dev/null 2>&1 || { echo "python3 no encontrado; omitido (opcional)"; exit 0; }; python3 benchmarks/compare.py

# Coverage across the full module.
coverage:
	go test -cover ./...
