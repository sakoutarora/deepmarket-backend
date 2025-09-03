// engine/indicators.go
package engine

import (
	"errors"
	"fmt"
	"log"

	domain "github.com/gulll/deepmarket/backtesting/domain"
)

func BuildRegistry() *Registry {
	reg := &Registry{
		Indicators: map[string]IndicatorSpec{},
		Functions:  map[string]FunctionSpec{},
	}

	// Close
	reg.Indicators["Close"] = IndicatorSpec{
		Category:    "Price",
		Description: "Closing price",
		Eval: func(ctx *EvalCtx, tf domain.Timeframe,
			_ map[string]float64, offset int, args ...Series) ([]float64, error) {

			cl := ctx.cache["close"]
			if offset == 0 {
				return cl, nil
			}
			// positive offset = shift back by N bars
			if offset < 0 || offset >= len(cl) {
				return nil, errors.New("bad offset")
			}
			out := make([]float64, len(cl))
			for i := range cl {
				j := i - offset
				if j >= 0 {
					out[i] = cl[j]
				} // leading zeros remain 0
			}
			return out, nil
		},
	}

	reg.Indicators["Open"] = IndicatorSpec{
		Category:    "Price",
		Description: "Open price",
		Eval: func(ctx *EvalCtx, tf domain.Timeframe,
			_ map[string]float64, offset int, args ...Series) ([]float64, error) {

			cl := ctx.cache["open"]
			if offset == 0 {
				return cl, nil
			}
			// positive offset = shift back by N bars
			if offset < 0 || offset >= len(cl) {
				return nil, errors.New("bad offset")
			}
			out := make([]float64, len(cl))
			for i := range cl {
				j := i - offset
				if j >= 0 {
					out[i] = cl[j]
				} // leading zeros remain 0
			}
			return out, nil
		},
	}

	reg.Indicators["High"] = IndicatorSpec{
		Category:    "Price",
		Description: "Open price",
		Eval: func(ctx *EvalCtx, tf domain.Timeframe,
			_ map[string]float64, offset int, args ...Series) ([]float64, error) {

			cl := ctx.cache["high"]
			if offset == 0 {
				return cl, nil
			}
			// positive offset = shift back by N bars
			if offset < 0 || offset >= len(cl) {
				return nil, errors.New("bad offset")
			}
			out := make([]float64, len(cl))
			for i := range cl {
				j := i - offset
				if j >= 0 {
					out[i] = cl[j]
				} // leading zeros remain 0
			}
			return out, nil
		},
	}

	reg.Indicators["Low"] = IndicatorSpec{
		Category:    "Price",
		Description: "Open price",
		Eval: func(ctx *EvalCtx, tf domain.Timeframe,
			_ map[string]float64, offset int, args ...Series) ([]float64, error) {

			cl := ctx.cache["low"]
			if offset == 0 {
				return cl, nil
			}
			// positive offset = shift back by N bars
			if offset < 0 || offset >= len(cl) {
				return nil, errors.New("bad offset")
			}
			out := make([]float64, len(cl))
			for i := range cl {
				j := i - offset
				if j >= 0 {
					out[i] = cl[j]
				} // leading zeros remain 0
			}
			return out, nil
		},
	}

	reg.Indicators["Time"] = IndicatorSpec{
		Category:    "Time",
		Description: "Time",
		Eval: func(ctx *EvalCtx, tf domain.Timeframe,
			_ map[string]float64, offset int, args ...Series) ([]float64, error) {

			cl := ctx.cache["time"]
			if offset == 0 {
				return cl, nil
			}
			// positive offset = shift back by N bars
			if offset < 0 || offset >= len(cl) {
				return nil, errors.New("bad offset")
			}
			out := make([]float64, len(cl))
			for i := range cl {
				j := i - offset
				if j >= 0 {
					out[i] = cl[j]
				} // leading zeros remain 0
			}
			return out, nil
		},
	}

	// ---------------------------------------------------------------------
	// Supertrend(period, mult)
	reg.Indicators["Supertrend"] = IndicatorSpec{
		Category:    "Trend",
		Description: "Supertrend (flexible input)",
		Params: []ArgSpec{
			{Name: "period", Type: "int", Req: true},
			{Name: "mult", Type: "float", Req: true},
		},
		Eval: func(ctx *EvalCtx, tf domain.Timeframe,
			params map[string]float64, offset int, args ...Series) ([]float64, error) {

			// Ensure we have High/Low/Close series
			var high, low, close Series
			if len(args) >= 3 {
				high, low, close = args[0], args[1], args[2]
			} else {
				log.Println("using default series for Supertrend")
				high = ctx.cache["high"]
				low = ctx.cache["low"]
				close = ctx.cache["close"]
			}

			n := len(close)
			if len(high) != n || len(low) != n {
				return nil, fmt.Errorf("series length mismatch")
			}

			// Build candles
			bars := make([]domain.Candle, n)
			for i := 0; i < n; i++ {
				bars[i] = domain.Candle{
					High:  high[i],
					Low:   low[i],
					Close: close[i],
				}
			}

			// Call Supertrend
			trend, _ := Supertrend(bars, int(params["period"]), params["mult"])
			return trend, nil
		},
	}

	// ---------------------------------------------------------------------
	// Function: SMA(expr, periods)
	reg.Functions["SMA"] = FunctionSpec{
		Category:    "Technical",
		Description: "Simple moving average",
		Params: []ArgSpec{
			{Name: "period", Type: "int", Req: true},
		},
		Eval: func(ctx *EvalCtx, params map[string]any, args ...Series) ([]float64, error) {
			if len(args) == 0 {
				return nil, errors.New("SMA requires an input series")
			}
			period := int(params["period"].(float64))
			return SMA(args[0], period), nil
		},
	}

	reg.Functions["EMA"] = FunctionSpec{
		Category:    "Technical",
		Description: "Exponential moving average",
		Params: []ArgSpec{
			{Name: "period", Type: "int", Req: true},
		},
		Eval: func(ctx *EvalCtx, params map[string]any, args ...Series) ([]float64, error) {
			if len(args) == 0 {
				return nil, errors.New("SMA requires an input series")
			}
			period := int(params["period"].(float64))
			return EMA(args[0], period), nil
		},
	}

	reg.Indicators["RSI"] = IndicatorSpec{
		Category:    "Technical",
		Description: "Relative Strength Indicator",
		Params: []ArgSpec{
			{Name: "period", Type: "int", Req: true},
		},
		Eval: func(ctx *EvalCtx, tf domain.Timeframe, params map[string]float64, offset int, args ...Series) ([]float64, error) {
			if len(args) == 0 {
				return nil, errors.New("RSI requires an input series")
			}
			period := int(params["period"])
			return RSI(args[0], period), nil
		},
	}

	return reg
}
