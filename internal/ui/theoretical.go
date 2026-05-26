package ui

import (
	"errors"
	"math"

	"github.com/Starbhein/DistCalc/internal/core/distributions"
)

// TheoreticalStats guarda los estadísticos teóricos de una distribución.
type TheoreticalStats struct {
	Avg      float64
	Variance float64
	StdDev   float64
}

// ComputeTheoreticalStats calcula media, varianza y desv. estándar teóricas.
func ComputeTheoreticalStats(distribution string, params []float64) (TheoreticalStats, error) {
	switch distribution {
	case "Binomial":
		if len(params) < 2 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Binomial")
		}
		p := params[0]
		n := params[1]
		avg := n * p
		variance := n * p * (1 - p)
		return TheoreticalStats{
			Avg:      avg,
			Variance: variance,
			StdDev:   math.Sqrt(variance),
		}, nil

	case "Poisson":
		if len(params) < 1 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Poisson")
		}
		lambda := params[0]
		return TheoreticalStats{
			Avg:      lambda,
			Variance: lambda,
			StdDev:   math.Sqrt(lambda),
		}, nil

	case "Hypergeométrica":
		if len(params) < 3 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Hypergeométrica")
		}
		N := params[0]
		M := params[1]
		n := params[2]
		avg := n * M / N
		variance := n * (M / N) * ((N - M) / N) * ((N - n) / (N - 1))
		return TheoreticalStats{
			Avg:      avg,
			Variance: variance,
			StdDev:   math.Sqrt(variance),
		}, nil

	case "Normal":
		if len(params) < 2 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Normal")
		}
		mu := params[0]
		sigma := params[1]
		return TheoreticalStats{
			Avg:      mu,
			Variance: sigma * sigma,
			StdDev:   sigma,
		}, nil

	case "Exponencial":
		if len(params) < 1 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Exponencial")
		}
		lambda := params[0]
		avg := 1.0 / lambda
		variance := 1.0 / (lambda * lambda)
		return TheoreticalStats{
			Avg:      avg,
			Variance: variance,
			StdDev:   avg,
		}, nil

	case "Exponencial (β)":
		if len(params) < 1 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Exponencial (β)")
		}
		beta := params[0]
		return TheoreticalStats{
			Avg:      beta,
			Variance: beta * beta,
			StdDev:   beta,
		}, nil

	case "Bernoulli":
		if len(params) < 1 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Bernoulli")
		}
		p := params[0]
		avg := p
		variance := p * (1.0 - p)
		return TheoreticalStats{
			Avg:      avg,
			Variance: variance,
			StdDev:   math.Sqrt(variance),
		}, nil

	case "Geométrica":
		if len(params) < 1 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Geométrica")
		}
		p := params[0]
		avg := 1.0 / p
		variance := (1.0 - p) / (p * p)
		return TheoreticalStats{
			Avg:      avg,
			Variance: variance,
			StdDev:   math.Sqrt(variance),
		}, nil

	case "Uniforme continua":
		if len(params) < 2 {
			return TheoreticalStats{}, errors.New("faltan parámetros para Uniforme continua")
		}
		a := params[0]
		b := params[1]
		avg := (a + b) / 2.0
		width := b - a
		variance := (width * width) / 12.0
		return TheoreticalStats{
			Avg:      avg,
			Variance: variance,
			StdDev:   math.Sqrt(variance),
		}, nil

	default:
		return TheoreticalStats{}, errors.New("distribución desconocida: " + distribution)
	}
}

// Probabilities guarda las probabilidades teóricas P(X=x), P(X≤x), P(X>x).
type Probabilities struct {
	PX  float64
	PLe float64
	PGt float64
}

// ComputeProbabilities calcula P(X=x), P(X≤x) y P(X>x) para la distribución dada.
// params incluye todos los parámetros de distribución incluyendo x al final.
func ComputeProbabilities(distribution string, params []float64) (Probabilities, error) {
	if len(params) < 1 {
		return Probabilities{}, errors.New("faltan parámetros")
	}

	// x siempre es el último parámetro
	x := params[len(params)-1]

	switch distribution {
	case "Binomial":
		if len(params) < 3 {
			return Probabilities{}, errors.New("faltan parámetros para Binomial")
		}
		p := params[0]
		n := int(params[1])
		k := int(math.Round(x))
		b, err := distributions.NewBinomial(n, p)
		if err != nil {
			return Probabilities{}, err
		}
		pmf, err := b.PMF(k)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := b.CDF(k)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pmf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Poisson":
		if len(params) < 2 {
			return Probabilities{}, errors.New("faltan parámetros para Poisson")
		}
		lambda := params[0]
		k := int(math.Round(x))
		p, err := distributions.NewPoisson(lambda)
		if err != nil {
			return Probabilities{}, err
		}
		pmf, err := p.PMF(k)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := p.CDF(k)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pmf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Hypergeométrica":
		if len(params) < 4 {
			return Probabilities{}, errors.New("faltan parámetros para Hypergeométrica")
		}
		N := int(params[0])
		M := int(params[1])
		n := int(params[2])
		k := int(math.Round(x))
		h, err := distributions.NewHypergeometric(M, N, n)
		if err != nil {
			return Probabilities{}, err
		}
		pmf, err := h.PMF(k)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := h.CDF(k)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pmf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Normal":
		if len(params) < 3 {
			return Probabilities{}, errors.New("faltan parámetros para Normal")
		}
		mu := params[0]
		sigma := params[1]
		n, err := distributions.NewNormal(mu, sigma)
		if err != nil {
			return Probabilities{}, err
		}
		pdf, err := n.PDF(x)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := n.CDF(x)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pdf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Exponencial":
		if len(params) < 2 {
			return Probabilities{}, errors.New("faltan parámetros para Exponencial")
		}
		lambda := params[0]
		el, err := distributions.NewExponentialLambda(lambda)
		if err != nil {
			return Probabilities{}, err
		}
		pdf, err := el.PDF(x)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := el.CDF(x)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pdf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Exponencial (β)":
		if len(params) < 2 {
			return Probabilities{}, errors.New("faltan parámetros para Exponencial (β)")
		}
		beta := params[0]
		eb, err := distributions.NewExponentialBetha(beta)
		if err != nil {
			return Probabilities{}, err
		}
		pdf, err := eb.PDF(x)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := eb.CDF(x)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pdf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Bernoulli":
		if len(params) < 2 {
			return Probabilities{}, errors.New("faltan parámetros para Bernoulli")
		}
		p := params[0]
		k := int(math.Round(x))
		b, err := distributions.NewBernoulli(p)
		if err != nil {
			return Probabilities{}, err
		}
		pmf, err := b.PMF(k)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := b.CDF(k)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pmf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Geométrica":
		if len(params) < 2 {
			return Probabilities{}, errors.New("faltan parámetros para Geométrica")
		}
		p := params[0]
		k := int(math.Round(x))
		g, err := distributions.NewGeometric(p)
		if err != nil {
			return Probabilities{}, err
		}
		pmf, err := g.PMF(k)
		if err != nil {
			return Probabilities{}, err
		}
		cdf, err := g.CDF(k)
		if err != nil {
			return Probabilities{}, err
		}
		return Probabilities{PX: pmf, PLe: cdf, PGt: 1 - cdf}, nil

	case "Uniforme continua":
		if len(params) < 3 {
			return Probabilities{}, errors.New("faltan parámetros para Uniforme continua")
		}
		a := params[0]
		b := params[1]
		if a >= b {
			return Probabilities{}, errors.New("a debe ser menor que b")
		}
		width := b - a
		var pdf, cdf float64
		if x < a {
			pdf = 0
			cdf = 0
		} else if x > b {
			pdf = 0
			cdf = 1
		} else {
			pdf = 1.0 / width
			cdf = (x - a) / width
		}
		return Probabilities{PX: pdf, PLe: cdf, PGt: 1 - cdf}, nil

	default:
		return Probabilities{}, errors.New("distribución desconocida: " + distribution)
	}
}
