// engine/evalctx.go
package engine

import (
	"errors"
	"math"

	"github.com/gulll/deepmarket/backtesting/domain"
)

type Series []float64
type BoolSeries []bool

// provides historical OHLCV for symbol+timeframe (aligned same length per tf)
type DataProvider interface {
	LoadOHLCV(symbol string, tf domain.Timeframe) ([]domain.Candle, error)
	AlignTo(baseTF domain.Timeframe, series Series, fromTF domain.Timeframe) (Series, error)
}

type EvalPolicy struct {
	// When comparing floats, optionally treat NaN as false (skip) instead of propagating.
	NaNIsFalse bool
}

type EvalCtx struct {
	Symbol string
	BaseTF domain.Timeframe
	Data   DataProvider
	Reg    *Registry

	// memoize computed indicator/function series by a key
	cache map[string]Series
	// memoize booleans
	bcache map[string]BoolSeries

	Policy EvalPolicy
}

func NewEvalCtx(sym string, baseTF domain.Timeframe, dp DataProvider, reg *Registry) *EvalCtx {
	return &EvalCtx{
		Symbol: sym, BaseTF: baseTF, Data: dp, Reg: reg,
		cache: map[string]Series{}, bcache: map[string]BoolSeries{},
		Policy: EvalPolicy{NaNIsFalse: true},
	}
}

func (ctx *EvalCtx) GetCache() map[string]Series { return ctx.cache }
func (ctx *EvalCtx) SetCache(cache map[string]Series) {
	ctx.cache = cache
	if ctx.bcache == nil {
		ctx.bcache = map[string]BoolSeries{}
	}
}

func eqLen(a, b Series) error {
	if len(a) != len(b) {
		return errors.New("series length mismatch")
	}
	return nil
}

func (ctx *EvalCtx) sanitizeBoolFromPair(l, r Series, f func(i int) bool) BoolSeries {
	out := make(BoolSeries, len(l))
	for i := range l {
		if math.IsNaN(l[i]) || math.IsNaN(r[i]) {
			out[i] = !ctx.Policy.NaNIsFalse // if NaNIsFalse, then false; else true
		} else {
			out[i] = f(i)
		}
	}
	return out
}
