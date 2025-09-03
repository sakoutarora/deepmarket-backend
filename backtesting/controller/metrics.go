package controller

import (
	"math"
)

// Sharpe Ratio = (mean(returns) - rf) / std(returns)
func CalcSharpe(returns []float64, rf float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	var sum, excessSum float64
	for _, r := range returns {
		sum += r
		excessSum += r - rf
	}
	meanExcess := excessSum / float64(len(returns))

	// std deviation
	var variance float64
	for _, r := range returns {
		diff := (r - rf) - meanExcess
		variance += diff * diff
	}
	std := math.Sqrt(variance / float64(len(returns)))
	if std == 0 {
		return 0
	}
	return meanExcess / std
}

// Sortino Ratio = (mean(returns) - rf) / downside deviation
func CalcSortino(returns []float64, rf float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	var excessSum float64
	var downsideVar float64
	n := float64(len(returns))
	for _, r := range returns {
		excess := r - rf
		excessSum += excess
		if excess < 0 {
			downsideVar += excess * excess
		}
	}
	meanExcess := excessSum / n
	if downsideVar == 0 {
		return 0
	}
	downsideDev := math.Sqrt(downsideVar / n)
	return meanExcess / downsideDev
}

// Calmar already correct, included for completeness
func CalcCalmar(cagr, maxDD float64) float64 {
	if maxDD == 0 {
		return 0
	}
	return cagr / math.Abs(maxDD)
}

// Omega Ratio = (sum of gains above threshold) / (sum of losses below threshold)
func CalcOmega(returns []float64, threshold float64) float64 {
	var gains, losses float64
	for _, r := range returns {
		excess := r - threshold
		if excess > 0 {
			gains += excess
		} else {
			losses += -excess
		}
	}
	if losses == 0 {
		if gains == 0 {
			return 0
		}
		return math.Inf(1)
	}
	return gains / losses
}

// Drawdowns: maxDD (biggest drop), avgDD (average depth), ulcer index (sqrt(mean(drawdown^2)))
func CalcDrawdowns(equity []float64) (maxDD, avgDD, ulcer float64) {
	if len(equity) == 0 {
		return 0, 0, 0
	}

	peak := equity[0]
	var ddSum, ddSqSum float64
	var ddCount int
	maxDD = 0

	for _, v := range equity {
		if v > peak {
			peak = v
		}
		dd := (v - peak) / peak
		if dd < 0 {
			ddSum += dd
			ddSqSum += dd * dd
			ddCount++
			if dd < maxDD {
				maxDD = dd
			}
		}
	}

	if ddCount > 0 {
		avgDD = ddSum / float64(ddCount)
		ulcer = math.Sqrt(ddSqSum / float64(ddCount))
	}
	return maxDD, avgDD, ulcer
}

// CAGR = (end/start)^(1/years) - 1
func CalcCAGR(start, end float64, years float64) float64 {
	if start <= 0 || end <= 0 || years <= 0 {
		return 0
	}
	return math.Pow(end/start, 1/years) - 1
}
