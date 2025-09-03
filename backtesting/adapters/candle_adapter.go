package adapters

import (
	"github.com/gulll/deepmarket/backtesting/domain"
	"github.com/gulll/deepmarket/backtesting/engine"
)

func CandlesToSeries(candles []domain.Candle) map[string]engine.Series {
	var ts, open, high, low, close, volume engine.Series

	for _, c := range candles {
		ts = append(ts, float64(c.Time.Unix())) // seconds since epoch
		// or: float64(c.Time.UnixMilli()) for ms precision

		open = append(open, c.Open)
		high = append(high, c.High)
		low = append(low, c.Low)
		close = append(close, c.Close)
		volume = append(volume, c.Volume)
	}

	return map[string]engine.Series{
		"time":   ts,
		"open":   open,
		"high":   high,
		"low":    low,
		"close":  close,
		"volume": volume,
	}
}
