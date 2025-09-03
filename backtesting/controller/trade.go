package controller

import (
	"time"

	"github.com/gulll/deepmarket/backtesting/domain"
)

type Trade struct {
	Direction     int
	EntryTime     time.Time
	EntryPrice    float64
	ExitTime      time.Time
	ExitPrice     float64
	Qty           int
	Open          bool
	HighWaterMark float64 // highest price seen (for long)
	LowWaterMark  float64 // lowest price seen (for short)
	ActiveTSL     float64 // the current trailing SL level
}

func NewTrade(entryTime time.Time, entryPrice float64, qty int, dir int) *Trade {
	return &Trade{EntryTime: entryTime, EntryPrice: entryPrice, Qty: qty, Direction: dir, Open: true}
}

func (t *Trade) Close(exitTime time.Time, exitPrice float64, reason string) domain.TradeLog {
	t.Open = false
	t.ExitTime = exitTime
	t.ExitPrice = exitPrice
	pnl := 0.0
	if t.Direction == 1 {
		pnl = (exitPrice - t.EntryPrice) * float64(t.Qty) // long PnL
	} else {
		pnl = (t.EntryPrice - exitPrice) * float64(t.Qty) // short PnL
	}

	return domain.TradeLog{
		EntryTime:   t.EntryTime,
		EntryPrice:  t.EntryPrice,
		ExitTime:    t.ExitTime,
		ExitPrice:   t.ExitPrice,
		ExitReason:  reason,
		Qty:         t.Qty,
		PnL:         pnl,
		HoldingBars: int(exitTime.Sub(t.EntryTime).Minutes()),
		Direction:   map[int]string{1: "long", -1: "short"}[t.Direction],
	}
}
