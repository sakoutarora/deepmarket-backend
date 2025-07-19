package options

import (
	"log"
	"math"

	"gonum.org/v1/gonum/optimize"
)

type OptionType string

const (
	Call OptionType = "call"
	Put  OptionType = "put"
)

// Black-Scholes price formula
func blackScholesPrice(S, K, T, r, sigma float64, optionType OptionType) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	if optionType == Call {
		return S*normCDF(d1) - K*math.Exp(-r*T)*normCDF(d2)
	} else {
		return K*math.Exp(-r*T)*normCDF(-d2) - S*normCDF(-d1)
	}
}

// Implied volatility using Brent's method
func ImpliedVolatility(marketPrice, S, K, T, r float64, optionType OptionType) float64 {
	problem := optimize.Problem{
		Func: func(x []float64) float64 {
			sigma := x[0]
			if sigma <= 0 {
				return math.Inf(1)
			}
			modelPrice := blackScholesPrice(S, K, T, r, sigma, optionType)
			return math.Pow(modelPrice-marketPrice, 2)
		},
	}
	result, err := optimize.Minimize(problem, []float64{0.2}, nil, nil)
	if err != nil {
		log.Printf("Error in IV calculation: %v", err)
		return math.NaN()
	}
	return result.X[0]
}

// Greeks

func Delta(S, K, T, r, sigma float64, optionType OptionType) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	if optionType == Call {
		return normCDF(d1)
	} else {
		return normCDF(d1) - 1
	}
}

func Gamma(S, K, T, r, sigma float64) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	return normPDF(d1) / (S * sigma * math.Sqrt(T))
}

func Vega(S, K, T, r, sigma float64) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	return S * normPDF(d1) * math.Sqrt(T)
}

func Theta(S, K, T, r, sigma float64, optionType OptionType) float64 {
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	d2 := d1 - sigma*math.Sqrt(T)

	term1 := -(S * normPDF(d1) * sigma) / (2 * math.Sqrt(T))

	if optionType == Call {
		term2 := r * K * math.Exp(-r*T) * normCDF(d2)
		return term1 - term2
	} else {
		term2 := r * K * math.Exp(-r*T) * normCDF(-d2)
		return term1 + term2
	}
}

func Rho(S, K, T, r, sigma float64, optionType OptionType) float64 {
	d2 := (math.Log(S/K) + (r-0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
	if optionType == Call {
		return K * T * math.Exp(-r*T) * normCDF(d2)
	} else {
		return -K * T * math.Exp(-r*T) * normCDF(-d2)
	}
}

// Standard Normal CDF
func normCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

// Standard Normal PDF
func normPDF(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}
