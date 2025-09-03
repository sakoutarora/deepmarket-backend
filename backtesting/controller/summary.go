package controller

import (
	"math"

	"github.com/gulll/deepmarket/backtesting/domain"
)

func ComputeSummary(trades []domain.TradeLog, equity []float64, startEquity float64) domain.BacktestSummary {
	if len(equity) == 0 {
		return domain.BacktestSummary{}
	}

	var grossProfit, grossLoss float64
	var wins, losses int
	var totalWinPnL, totalLossPnL float64
	var consecWins, consecLosses, maxConsecWins, maxConsecLosses int
	var totalHoldBars int

	// collect returns (equity % change)
	returns := make([]float64, 0, len(equity)-1)
	for i := 1; i < len(equity); i++ {
		ret := (equity[i] - equity[i-1]) / equity[i-1]
		returns = append(returns, ret)
	}

	for _, t := range trades {
		if t.PnL > 0 {
			grossProfit += t.PnL
			totalWinPnL += t.PnL
			wins++
			consecWins++
			if consecWins > maxConsecWins {
				maxConsecWins = consecWins
			}
			consecLosses = 0
		} else if t.PnL < 0 {
			grossLoss += t.PnL
			totalLossPnL += t.PnL
			losses++
			consecLosses++
			if consecLosses > maxConsecLosses {
				maxConsecLosses = consecLosses
			}
			consecWins = 0
		} else {
			// break streak on breakeven
			consecWins = 0
			consecLosses = 0
		}

		totalHoldBars += t.HoldingBars
	}

	netProfit := equity[len(equity)-1] - startEquity

	// Profit Factor
	var profitFactor float64
	if grossLoss != 0 {
		profitFactor = grossProfit / math.Abs(grossLoss)
	}

	// Expectancy
	var expectancy float64
	if len(trades) > 0 {
		expectancy = netProfit / float64(len(trades))
	}

	// Risk-adjusted metrics
	sharpe := CalcSharpe(returns, 0)
	sortino := CalcSortino(returns, 0)
	maxDD, avgDD, ulcer := CalcDrawdowns(equity)
	years := equityDurationYears(trades) // helper below
	cagr := CalcCAGR(startEquity, equity[len(equity)-1], years)
	calmar := CalcCalmar(cagr, maxDD)
	omega := CalcOmega(returns, 0)

	// Trade quality
	var winRate, avgWin, avgLoss, rrRatio float64
	if len(trades) > 0 {
		winRate = float64(wins) / float64(len(trades))
	}
	if wins > 0 {
		avgWin = totalWinPnL / float64(wins)
	}
	if losses > 0 {
		avgLoss = totalLossPnL / float64(losses)
	}
	if avgLoss != 0 {
		rrRatio = avgWin / math.Abs(avgLoss)
	}

	// Recovery Factor = Net Profit / |Max Drawdown|
	var recovery float64
	if maxDD != 0 {
		recovery = netProfit / math.Abs(maxDD*startEquity)
	}

	// Exposure & turnover
	var avgHoldBars float64
	if len(trades) > 0 {
		avgHoldBars = float64(totalHoldBars) / float64(len(trades))
	}
	// simple proxies (customize later)
	exposureRatio := avgHoldBars / float64(len(equity))
	turnoverRatio := float64(len(trades)) / float64(len(equity))

	return domain.BacktestSummary{
		TotalTrades:      len(trades),
		NetProfit:        netProfit,
		GrossProfit:      grossProfit,
		GrossLoss:        grossLoss,
		ProfitFactor:     profitFactor,
		Expectancy:       expectancy,
		SharpeRatio:      sharpe,
		SortinoRatio:     sortino,
		CalmarRatio:      calmar,
		OmegaRatio:       omega,
		MaxDrawdown:      maxDD,
		AvgDrawdown:      avgDD,
		RecoveryFactor:   recovery,
		UlcerIndex:       ulcer,
		WinRate:          winRate,
		AvgWin:           avgWin,
		AvgLoss:          avgLoss,
		RiskRewardRatio:  rrRatio,
		MaxConsecWins:    maxConsecWins,
		MaxConsecLosses:  maxConsecLosses,
		CAGR:             cagr,
		EquityVolatility: stdDev(returns),
		// For skew/kurt you may add proper moment calcs
		Skewness:      skew(returns),
		Kurtosis:      kurtosis(returns),
		AvgHoldBars:   avgHoldBars,
		ExposureRatio: exposureRatio,
		TurnoverRatio: turnoverRatio,
		Trades:        trades,
	}
}

// --- Helpers --- //

func equityDurationYears(trades []domain.TradeLog) float64 {
	if len(trades) == 0 {
		return 1
	}
	start := trades[0].EntryTime
	end := trades[len(trades)-1].ExitTime
	days := end.Sub(start).Hours() / 24.0
	return days / 365.0
}

func stdDev(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range x {
		mean += v
	}
	mean /= float64(len(x))

	var sum float64
	for _, v := range x {
		diff := v - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(x)))
}

func skew(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	n := float64(len(x))
	mean := 0.0
	for _, v := range x {
		mean += v
	}
	mean /= n

	var m2, m3 float64
	for _, v := range x {
		d := v - mean
		m2 += d * d
		m3 += d * d * d
	}
	m2 /= n
	m3 /= n

	if m2 == 0 {
		return 0
	}
	return m3 / math.Pow(m2, 1.5)
}

func kurtosis(x []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	n := float64(len(x))
	mean := 0.0
	for _, v := range x {
		mean += v
	}
	mean /= n

	var m2, m4 float64
	for _, v := range x {
		d := v - mean
		m2 += d * d
		m4 += d * d * d * d
	}
	m2 /= n
	m4 /= n

	if m2 == 0 {
		return 0
	}
	return m4 / (m2 * m2)
}
