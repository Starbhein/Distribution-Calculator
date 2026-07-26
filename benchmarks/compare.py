#!/usr/bin/env python3
"""Algorithm-class evidence for the Distribution Calculator math cores.

Two independent parts, both pure Python (stdlib only, no third-party deps):

1. Accuracy table — exact rational references (math.comb + fractions.Fraction)
   vs the float64 row recurrences used by the Go distmath package, ported
   here line-for-line. Reports max absolute/relative drift.

2. Scaling sweep — float64 only (no Fraction). Contrasts ALGORITHM CLASSES,
   never languages:
   - Hypergeometric support scan: O(S) mode-anchored row recurrence vs the
     MANDATORY naive per-point closed form (direct multiplicative binomial
     product, O(k) per point -> O(S^2) total). An lgamma per-point variant
     would be linear-with-big-constant and would NOT show the class gap.
   - Geometric fill: O(1) inverse-CDF draw (k = ceil(log(u)/log1p(-p))) vs
     the O(1/p) expected-trials loop.

This script NEVER runs, parses, or prints Go benchmark numbers, and never
ranks languages. Go timings come from `make bench`; correlate by hand.
Correctness cross-checks against the Go pinned test values are MANUAL:
compare the exact values below against benchmarks/README.md.
"""

import math
import platform
import statistics
import sys
import time
from fractions import Fraction

# distmath.EpsilonSignificantValue (design §2.3): single truncation epsilon
# for every recurrence core.
_EPSILON_SIGNIFICANT = 1e-18

# Deterministic local PRNG (64-bit LCG + 53-bit mapping) so the module keeps
# the pinned import set (no `random`) and every sweep run is reproducible.
_LCG_MULT = 6364136223846793005
_LCG_INC = 1442695040888963407
_MASK64 = (1 << 64) - 1
_U01_SCALE = 1.0 / (1 << 53)


def _next_u01(state):
    """Advance the LCG; return (new_state, u) with u in (0, 1].

    The +0.5 offset keeps u off 0, so log(u) is always finite without a
    zero guard (mirrors the Go FillGeometric u == 0 guard). It does NOT
    keep u off 1: (2^53 - 1 + 0.5) / 2^53 rounds to exactly 1.0 in float64
    (round-half-to-even), so u == 1.0 is reachable. Then log(1.0) == 0
    gives k == 0 in the inverse-CDF draw; the k >= 1 clamp in
    geometric_fill_inverse_cdf maps it to k = 1, the correct boundary draw.
    """
    state = (state * _LCG_MULT + _LCG_INC) & _MASK64
    return state, ((state >> 11) + 0.5) * _U01_SCALE


# --------------------------------------------------------------------------
# Exact rational references (accuracy table only — never used in the sweep).
# --------------------------------------------------------------------------

def hypergeometric_exact_row(m, n_sample, n):
    """Exact hypergeometric PMF row over the support as Fractions:
    C(M,k) * C(N-M, n-k) / C(N,n) with one shared denominator."""
    start_k = max(0, n_sample - (n - m))
    max_k = min(n_sample, m)
    den = math.comb(n, n_sample)
    row = [Fraction(math.comb(m, k) * math.comb(n - m, n_sample - k), den)
           for k in range(start_k, max_k + 1)]
    return row, start_k, max_k


def binomial_exact_row(n, a, b):
    """Exact binomial PMF row over [0, n] for rational p = a/b:
    C(n,k) * a^k * (b-a)^(n-k) / b^n with one shared denominator."""
    den = b ** n
    q = b - a
    return [Fraction(math.comb(n, k) * a ** k * q ** (n - k), den)
            for k in range(n + 1)]


# --------------------------------------------------------------------------
# float64 recurrence ports (same algorithms as the Go distmath core).
# --------------------------------------------------------------------------

def _log_hypergeometric_pmf(m, n_sample, n, k):
    return (math.lgamma(m + 1) - math.lgamma(k + 1) - math.lgamma(m - k + 1)
            + math.lgamma(n - m + 1) - math.lgamma(n_sample - k + 1)
            - math.lgamma(n - m - n_sample + k + 1)
            + math.lgamma(n_sample + 1) + math.lgamma(n - n_sample + 1)
            - math.lgamma(n + 1))


def _hypergeometric_ratio(m, n_sample, n, k):
    """P(k)/P(k-1) = (n-k+1)(M-k+1) / (k(N-M-n+k)); callers keep startK < k
    <= maxK so numerator and denominator are strictly positive."""
    return ((n_sample - k + 1) * (m - k + 1)) / (k * (n - m - n_sample + k))


def hypergeometric_pmf_row_float(m, n_sample, n):
    """float64 port of distmath.HypergeometricPMFRow: one lgamma seed at the
    clamped mode plus an O(range) ratio walk, normalized to sum 1."""
    start_k = max(0, n_sample - (n - m))
    max_k = min(n_sample, m)
    mode = ((n_sample + 1) * (m + 1)) // (n + 2)
    mode = min(max(mode, start_k), max_k)
    row = [0.0] * (max_k - start_k + 1)
    row[mode - start_k] = math.exp(_log_hypergeometric_pmf(m, n_sample, n, mode))
    for k in range(mode, start_k, -1):
        row[k - 1 - start_k] = row[k - start_k] / _hypergeometric_ratio(m, n_sample, n, k)
    for k in range(mode + 1, max_k + 1):
        row[k - start_k] = row[k - 1 - start_k] * _hypergeometric_ratio(m, n_sample, n, k)
    total = sum(row)
    return [v / total for v in row], start_k, max_k


def _log_binomial_pmf(n, p, k):
    return (math.lgamma(n + 1) - math.lgamma(k + 1) - math.lgamma(n - k + 1)
            + k * math.log(p) + (n - k) * math.log(1.0 - p))


def _binomial_cdf_core(n, p, hi, row=None):
    """float64 port of distmath.binomialCDFCore: anchor at min(int(n*p), hi),
    pre-divided ratios, two-direction gated walk. When row is given, every
    visited relative PMF is stored (materialized shape)."""
    pre_i = (1.0 - p) / p
    pre_r = p / (1.0 - p)
    max_value = min(int(n * p), hi)
    if row is not None:
        row[max_value] = 1.0
    total = 1.0
    cumulative = 1.0
    i = max_value - 1
    while i >= 0 and cumulative >= total * _EPSILON_SIGNIFICANT:
        cumulative *= pre_i * ((i + 1) / (n - i))
        total += cumulative
        if row is not None:
            row[i] = cumulative
        i -= 1
    cumulative = 1.0
    i = max_value + 1
    while i <= hi and cumulative >= total * _EPSILON_SIGNIFICANT:
        cumulative *= pre_r * ((n - i + 1) / i)
        total += cumulative
        if row is not None:
            row[i] = cumulative
        i += 1
    return max_value, total


def binomial_pmf_row_float(n, p):
    """float64 port of distmath.BinomialPMFRow: materialized core, normalized
    by the accumulated relative-PMF sum."""
    row = [0.0] * (n + 1)
    _, total = _binomial_cdf_core(n, p, n, row)
    return [v / total for v in row]


def binomial_cdf_float(n, p, k):
    """float64 port of distmath.BinomialCDF (pointwise, allocation-free)."""
    if k < 0:
        return 0.0
    if k >= n:
        return 1.0
    max_value, total = _binomial_cdf_core(n, p, k)
    return math.exp(_log_binomial_pmf(n, p, max_value)) * total


# --------------------------------------------------------------------------
# Accuracy table.
# --------------------------------------------------------------------------

def _max_drift(exact_row, float_row):
    """Max absolute drift over the support; max relative drift restricted to
    points with exact mass >= 1e-15 (relative error on far-tail dust is noise,
    not signal)."""
    max_abs = 0.0
    max_rel = 0.0
    for exact, got in zip(exact_row, float_row):
        want = float(exact)
        diff = abs(got - want)
        max_abs = max(max_abs, diff)
        if want >= 1e-15:
            max_rel = max(max_rel, diff / want)
    return max_abs, max_rel


def _print_spot(label, exact, got):
    diff = abs(got - float(exact))
    rel = diff / float(exact) if exact else 0.0
    shown = str(exact)
    if len(shown) > 24:
        # Big-int Fractions (n=1000 case) print thousands of digits; show the
        # float value instead — benchmarks/README.md lists the closed forms.
        shown = "~%.17g (bigint fraction)" % float(exact)
    print("  %-34s exact %-24s float %.17g  abs %.2e rel %.2e"
          % (label, shown, got, diff, rel))


def _accuracy_hypergeometric(m, n_sample, n, spots):
    exact_row, start_k, max_k = hypergeometric_exact_row(m, n_sample, n)
    float_row, _, _ = hypergeometric_pmf_row_float(m, n_sample, n)
    print("Hypergeometric M=%d N=%d n=%d — support [%d, %d]"
          % (m, n, n_sample, start_k, max_k))
    for k in spots:
        _print_spot("PMF(%d)" % k, exact_row[k - start_k], float_row[k - start_k])
    max_abs, max_rel = _max_drift(exact_row, float_row)
    # CDF spots via the normalized cumsum (distmath.HypergeometricCDFTable
    # shape: forward walk, normalize, last entry forced to exactly 1).
    exact_cdf = sum(exact_row[:max_k - start_k])  # up to k = max_k - 1
    float_cdf = sum(float_row[:max_k - start_k])
    _print_spot("CDF(%d)" % (max_k - 1), exact_cdf, float_cdf)
    print("  max abs drift %.3e | max rel drift (mass >= 1e-15) %.3e | row sums: exact %s float %.17g"
          % (max_abs, max_rel, sum(exact_row), sum(float_row)))
    print()


def _accuracy_binomial(n, a, b, spots, cdf_spots):
    p = a / b
    exact_row = binomial_exact_row(n, a, b)
    float_row = binomial_pmf_row_float(n, p)
    print("Binomial n=%d p=%g — support [0, %d]" % (n, p, n))
    for k in spots:
        _print_spot("PMF(%d)" % k, exact_row[k], float_row[k])
    for k in cdf_spots:
        exact_cdf = sum(exact_row[:k + 1])
        _print_spot("CDF(%d)" % k, exact_cdf, binomial_cdf_float(n, p, k))
    max_abs, max_rel = _max_drift(exact_row, float_row)
    print("  max abs drift %.3e | max rel drift (mass >= 1e-15) %.3e | row sums: exact %s float %.17g"
          % (max_abs, max_rel, sum(exact_row), sum(float_row)))
    print()


# --------------------------------------------------------------------------
# Scaling sweep (float64 only — no Fraction anywhere below).
# --------------------------------------------------------------------------

def hypergeometric_naive_row(m, n_sample, n):
    """MANDATORY naive per-point closed form: each PMF is a direct
    multiplicative product of the three binomial factors, O(k) per point ->
    O(S^2) total. The three C(.,.) factors are evaluated as ONE interleaved
    product so intermediates stay finite at N=1600 (separate float C(.,.)
    values overflow float64 past ~1e308); the per-point cost stays O(k), so
    the O(S^2) class contrast is unchanged."""
    start_k = max(0, n_sample - (n - m))
    max_k = min(n_sample, m)
    row = []
    for k in range(start_k, max_k + 1):
        v = 1.0
        for i in range(1, n_sample + 1):
            if i <= k:
                v *= (m - k + i) / i
            if i <= n_sample - k:
                v *= (n - m - n_sample + k + i) / i
            v /= (n - n_sample + i) / i
        row.append(v)
    return row


def geometric_fill_inverse_cdf(m, p, state):
    """O(1) per sample: k = ceil(log(u)/log1p(-p)) — one PRNG draw each."""
    log1mp = math.log1p(-p)
    total = 0.0
    for _ in range(m):
        state, u = _next_u01(state)
        k = math.ceil(math.log(u) / log1mp)
        # u == 1.0 is reachable (see _next_u01): log(1.0) == 0 -> k == 0,
        # and the clamp maps it to k = 1 — no crash, no silent corruption.
        total += k if k >= 1 else 1
    return state, total


def geometric_fill_trials(m, p, state):
    """O(1/p) expected trials per sample: Bernoulli draws until success.
    The LCG is inlined to keep the worst leg (m/p = 2e7 iterations) cheap."""
    total = 0
    for _ in range(m):
        k = 0
        while True:
            state = (state * _LCG_MULT + _LCG_INC) & _MASK64
            k += 1
            if ((state >> 11) + 0.5) * _U01_SCALE < p:
                break
        total += k
    return state, total


def _median_seconds(func, reps):
    """Median wall time over reps; also returns the LAST run's result so
    callers can report sanity values without re-running expensive legs."""
    times = []
    result = None
    for _ in range(reps):
        start = time.perf_counter()
        result = func()
        times.append(time.perf_counter() - start)
    return statistics.median(times), result


def _log_log_slope(xs, ys):
    """Least-squares slope of log(y) vs log(x) — the measured growth
    exponent (stdlib-only math)."""
    lx = [math.log(x) for x in xs]
    ly = [math.log(y) for y in ys]
    mx = sum(lx) / len(lx)
    my = sum(ly) / len(ly)
    num = sum((x - mx) * (y - my) for x, y in zip(lx, ly))
    den = sum((x - mx) ** 2 for x in lx)
    return num / den


def _sweep_hypergeometric(ns, reps):
    print("Hypergeometric support scan: O(S) row recurrence vs naive O(S^2) per-point closed form")
    print("  %-6s %-6s %-6s %-5s %16s %16s" % ("N", "M", "n", "S", "row ms", "naive ms"))
    sizes, row_ms, naive_ms = [], [], []
    for n in ns:
        m, n_sample = n // 2, n // 4
        start_k = max(0, n_sample - (n - m))
        max_k = min(n_sample, m)
        support = max_k - start_k + 1
        t_row, _ = _median_seconds(lambda: hypergeometric_pmf_row_float(m, n_sample, n), reps)
        t_nv, _ = _median_seconds(lambda: hypergeometric_naive_row(m, n_sample, n), reps)
        sizes.append(support)
        row_ms.append(t_row * 1e3)
        naive_ms.append(t_nv * 1e3)
        print("  %-6d %-6d %-6d %-5d %16.4f %16.4f" % (n, m, n_sample, support, t_row * 1e3, t_nv * 1e3))
    print("  log-log slope vs S: row %.2f (expect ~1) | naive %.2f (expect ~2)"
          % (_log_log_slope(sizes, row_ms), _log_log_slope(sizes, naive_ms)))
    print()


def _sweep_geometric(m, ps, reps):
    print("Geometric fill, m=%d samples: O(1) inverse-CDF vs O(1/p) expected-trials loop" % m)
    print("  %-9s %-8s %16s %16s %22s" % ("p", "1/p", "inverse-CDF ms", "trials ms", "mean k (inv/trials)"))
    inv_p, inv_ms, tr_ms = [], [], []
    for p in ps:
        state = 42  # fixed seed: reproducible sanity means
        t_inv, (_, sum_inv) = _median_seconds(lambda: geometric_fill_inverse_cdf(m, p, state), reps)
        t_tr, (_, sum_tr) = _median_seconds(lambda: geometric_fill_trials(m, p, state), reps)
        inv_p.append(1.0 / p)
        inv_ms.append(t_inv * 1e3)
        tr_ms.append(t_tr * 1e3)
        print("  %-9g %-8g %16.4f %16.4f %13.2f / %-7.2f"
              % (p, 1.0 / p, t_inv * 1e3, t_tr * 1e3, sum_inv / m, sum_tr / m))
    print("  log-log slope vs 1/p: inverse-CDF %.2f (expect ~0) | trials %.2f (expect ~1)"
          % (_log_log_slope(inv_p, inv_ms), _log_log_slope(inv_p, tr_ms)))
    print()


def main():
    hyper_ns = [200, 400, 800, 1600]
    geo_ps = [0.1, 0.01, 0.001, 0.0001]
    geo_m = 2000
    hyper_reps = 5  # cheap legs: extra reps steady the small-size timings
    geo_reps = 3    # worst leg is m/p = 2e7 iterations; 3 is the design floor

    print("compare.py — algorithm-class evidence (accuracy + scaling sweep)")
    print("python   : %s" % sys.version.split()[0])
    print("platform : %s" % platform.platform())
    print("method   : accuracy = exact rational (math.comb + fractions.Fraction) vs the")
    print("           float64 row recurrence ported from the Go distmath core.")
    print("           sweep = float64 only; contrasts algorithm CLASSES, both in Python.")
    print("           timings = statistics.median of reps (hypergeometric %d, geometric %d)."
          % (hyper_reps, geo_reps))
    print("sizes    : accuracy — hypergeometric M=3 N=12 n=4 (pinned) and N=50 M=20 n=10;")
    print("           binomial n=10 p=0.1 (pinned) and n=1000 p=0.9 (benchmark case).")
    print("           sweep — hypergeometric N in %s (M=N/2, n=N/4);" % hyper_ns)
    print("           geometric m=%d, p in %s." % (geo_m, geo_ps))
    print("scope    : NEVER runs or parses Go output, NEVER ranks languages.")
    print("           Go timings come from `make bench`; correlate by hand.")
    print("           Cross-check the exact values against benchmarks/README.md.")
    print()

    print("== Accuracy: exact rational reference vs float64 recurrence ==")
    print()
    _accuracy_hypergeometric(3, 4, 12, spots=(0, 1, 2, 3))
    _accuracy_hypergeometric(20, 10, 50, spots=(0, 4, 10))
    _accuracy_binomial(10, 1, 10, spots=(0, 1, 2), cdf_spots=(2,))
    _accuracy_binomial(1000, 9, 10, spots=(900,), cdf_spots=(999,))

    print("== Scaling sweep (float64, no Fraction) ==")
    print()
    _sweep_hypergeometric(hyper_ns, hyper_reps)
    _sweep_geometric(geo_m, geo_ps, geo_reps)


if __name__ == "__main__":
    main()
