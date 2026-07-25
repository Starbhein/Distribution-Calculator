package ui

import (
	"errors"

	"github.com/Starbhein/DistCalc/internal/core/distributions/registry"
)

// TheoreticalStats guarda los estadísticos teóricos de una distribución.
type TheoreticalStats struct {
	Avg      float64
	Variance float64
	StdDev   float64
}

// ComputeTheoreticalStats calcula media, varianza y desv. estándar teóricas.
// Delegates to the distribution structs' own methods through the registry
// (design §1.3, spec §3) — no formulas live here anymore.
func ComputeTheoreticalStats(distribution string, params []float64) (TheoreticalStats, error) {
	spec, ok := registry.ByName(distribution)
	if !ok {
		return TheoreticalStats{}, errors.New("distribución desconocida: " + distribution)
	}
	avg, variance, stdDev, err := registry.TheoreticalStats(spec, params)
	if err != nil {
		return TheoreticalStats{}, err
	}
	return TheoreticalStats{
		Avg:      avg,
		Variance: variance,
		StdDev:   stdDev,
	}, nil
}

// Probabilities guarda las probabilidades teóricas P(X=x), P(X≤x), P(X>x).
type Probabilities struct {
	PX  float64
	PLe float64
	PGt float64
}

// ComputeProbabilities calcula P(X=x), P(X≤x) y P(X>x) para la distribución dada.
// params incluye todos los parámetros de distribución incluyendo x al final.
// Dispatch goes through the registry (design §1.3, spec §3).
func ComputeProbabilities(distribution string, params []float64) (Probabilities, error) {
	if len(params) < 1 {
		return Probabilities{}, errors.New("faltan parámetros")
	}

	// x siempre es el último parámetro
	x := params[len(params)-1]

	spec, ok := registry.ByName(distribution)
	if !ok {
		return Probabilities{}, errors.New("distribución desconocida: " + distribution)
	}
	px, ple, pgt, err := registry.Probabilities(spec, params[:len(params)-1], x)
	if err != nil {
		return Probabilities{}, err
	}
	return Probabilities{PX: px, PLe: ple, PGt: pgt}, nil
}
