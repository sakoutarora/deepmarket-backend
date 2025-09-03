package controller

import (
	"github.com/gulll/deepmarket/backtesting/domain"
	"github.com/gulll/deepmarket/backtesting/engine"
)

// TODO: things lacking here dynamic position sizing
func RunBacktest(req domain.BacktestReq, sym string, ctx *engine.EvalCtx, rt *engine.Runtime,
	entryPlan *engine.Plan, exitPlan *engine.Plan, ohlc []domain.Candle) ([]domain.TradeLog, []bool, []float64, error) {

	// Entry signals
	entrySer, err := rt.ExecPlan(entryPlan)
	if err != nil {
		return nil, nil, nil, err
	}

	exitSeries := make([]bool, len(ohlc))

	if exitPlan != nil {
		exitSeries, err = rt.ExecPlan(exitPlan)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	exitChecker := &ExitChecker{
		StopLoss:    req.StopLoss,
		TakeProfit:  req.TakeProfit,
		TrailingSL:  req.TrailingSL,
		HoldingBars: req.HoldingPeriod,
		Intraday:    req.Intraday,
	}

	var enrtyDirecction int
	if req.Direction == "long" {
		enrtyDirecction = 1
	} else {
		enrtyDirecction = -1
	}

	// Trade loop
	var trades []domain.TradeLog
	var activeTrade *Trade
	equity := []float64{}
	capital := float64(req.Capital) // configurable base

	for i, bar := range ohlc {
		price := bar.Close
		barTime := bar.Time

		// Close trade if open
		if activeTrade != nil && activeTrade.Open {
			exit, reason := exitChecker.CheckExit(activeTrade, price, barTime, i)
			if !exit {
				if i < len(exitSeries) && exitSeries[i] {
					exit, reason = true, "ExitCondition"
				}
			}
			if !exit {
				if ok, reason2 := exitChecker.CheckIntradayExit(barTime); ok {
					exit, reason = true, reason2
				}
			}
			if exit {
				log := activeTrade.Close(barTime, price, reason)
				trades = append(trades, log)
				capital += log.PnL
				activeTrade = nil
			}
		}

		// Entry
		if activeTrade == nil && i < len(entrySer) && entrySer[i] {
			if exitChecker.AllowEntry(barTime) {
				activeTrade = NewTrade(barTime, price, req.Quantity, enrtyDirecction)
			}
		}

		equity = append(equity, capital)
	}

	// Close leftover
	if activeTrade != nil && activeTrade.Open {
		last := ohlc[len(ohlc)-1]
		log := activeTrade.Close(last.Time, last.Close, "EndOfBacktest")
		trades = append(trades, log)
		capital += log.PnL
	}

	return trades, entrySer, equity, nil
}
