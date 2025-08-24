// domain/registry.go
package engine

import domain "github.com/gulll/deepmarket/domain/backtesting"

type ArgSpec struct {
	Name string
	Type string
	Req  bool
}

type IndicatorSpec struct {
	Category    string
	Description string
	Params      []ArgSpec // e.g. period:int, mult:float
	// Eval returns a series for the requested timeframe
	Eval func(ctx *EvalCtx, tf domain.Timeframe, params map[string]float64, offset int) ([]float64, error)
}

type FunctionSpec struct {
	Category    string
	Description string
	Params      []ArgSpec
	// Eval returns a single series derived from expression(s) or numbers
	Eval func(ctx *EvalCtx, params map[string]any) ([]float64, error)
}

type Registry struct {
	Indicators map[string]IndicatorSpec
	Functions  map[string]FunctionSpec
}
