// engine/indicators.go
package engine

import (
	"errors"

	domain "github.com/gulll/deepmarket/domain/backtesting"
)

func BuildRegistry() *Registry {
	reg := &Registry{
		Indicators: map[string]IndicatorSpec{},
		Functions:  map[string]FunctionSpec{},
	}
	// Close
	reg.Indicators["Close"] = IndicatorSpec{
		Category: "Price", Description: "Closing price",
		Eval: func(ctx *EvalCtx, tf domain.Timeframe, _ map[string]float64, offset int) ([]float64, error) {
			data, err := ctx.Data.LoadOHLCV(ctx.Symbol, tf)
			if err != nil {
				return nil, err
			}
			cl := data["close"]
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
	// Supertrend(period, mult)
	reg.Indicators["Supertrend"] = IndicatorSpec{
		Category: "Trend", Description: "Supertrend",
		Params: []ArgSpec{
			{Name: "period", Type: "int", Req: true},
			{Name: "mult", Type: "float", Req: true},
		},
		Eval: func(ctx *EvalCtx, tf domain.Timeframe, params map[string]float64, offset int) ([]float64, error) {
			// implement (use ATR and hl2 envelopes). For brevity here, return Close as placeholder.
			data, err := ctx.Data.LoadOHLCV(ctx.Symbol, tf)
			if err != nil {
				return nil, err
			}
			st := data["close"] // replace with actual ST calc
			return st, nil
		},
	}
	// Function: sma(expression, periods)
	reg.Functions["sma"] = FunctionSpec{
		Category: "Technical", Description: "Simple moving average",
		Params: []ArgSpec{
			{Name: "expression", Type: "expr", Req: true},
			{Name: "periods", Type: "int", Req: true},
		},
		Eval: func(ctx *EvalCtx, params map[string]any) ([]float64, error) {
			// For brevity, expect expression is constant or not used. Extend to accept nested ExprNode.
			pN := int(params["periods"].(float64))
			// demo: return flat zeros of base length
			data, err := ctx.Data.LoadOHLCV(ctx.Symbol, ctx.BaseTF)
			if err != nil {
				return nil, err
			}
			out := make([]float64, len(data["close"]))
			_ = pN
			return out, nil
		},
	}
	return reg
}
