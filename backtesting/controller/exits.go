package controller

import (
	"time"

	"github.com/gulll/deepmarket/backtesting/domain"
)

type ExitChecker struct {
	StopLoss    float64
	TakeProfit  float64
	TrailingSL  float64
	HoldingBars *int
	Intraday    *domain.IntradayRule
}

// Basic exit conditions (price/risk/time)
func (ec *ExitChecker) CheckExit(trade *Trade, price float64, barTime time.Time, barIndex int) (bool, string) {
	// Update trailing SL
	if ec.TrailingSL > 0 {
		if trade.Direction > 0 { // Long
			if price > trade.HighWaterMark {
				trade.HighWaterMark = price
				trade.ActiveTSL = trade.HighWaterMark * (1 - ec.TrailingSL/100)
			}
			if price <= trade.ActiveTSL {
				return true, "TrailingStop"
			}
		} else { // Short
			if price < trade.LowWaterMark {
				trade.LowWaterMark = price
				trade.ActiveTSL = trade.LowWaterMark * (1 + ec.TrailingSL/100)
			}
			if price >= trade.ActiveTSL {
				return true, "TrailingStop"
			}
		}
	}

	// Fixed SL
	if ec.StopLoss > 0 && price <= trade.EntryPrice*(1-ec.StopLoss/100) {
		return true, "StopLoss"
	}

	// TP
	if ec.TakeProfit > 0 && price >= trade.EntryPrice*(1+ec.TakeProfit/100) {
		return true, "TakeProfit"
	}

	// Holding period
	if ec.HoldingBars != nil && barIndex >= *ec.HoldingBars {
		return true, "MaxHoldingPeriod"
	}

	return false, ""
}

func (ec *ExitChecker) AllowEntry(barTime time.Time) bool {
	if ec.Intraday == nil || !ec.Intraday.Enabled || ec.Intraday.StartTime == "" {
		return true
	}
	start, _ := time.Parse("15:04", ec.Intraday.StartTime)
	if barTime.Hour() < start.Hour() ||
		(barTime.Hour() == start.Hour() && barTime.Minute() < start.Minute()) {
		return false
	}
	return true
}

func (ec *ExitChecker) CheckIntradayExit(barTime time.Time) (bool, string) {
	if ec.Intraday != nil && ec.Intraday.Enabled && ec.Intraday.ExitTime != "" {
		exit, _ := time.Parse("15:04", ec.Intraday.ExitTime)
		if barTime.Hour() > exit.Hour() ||
			(barTime.Hour() == exit.Hour() && barTime.Minute() >= exit.Minute()) {
			return true, "IntradayExit"
		}
	}
	return false, ""
}
