package handlers

import (
	domain "github.com/gulll/deepmarket/domain/backtesting"
	engine "github.com/gulll/deepmarket/engine/backtesting"

	"github.com/gofiber/fiber/v2"
)

type BacktestReq struct {
	Symbols    []string           `json:"symbols"`
	BaseTF     domain.Timeframe   `json:"base_timeframe"`
	Conditions []domain.Condition `json:"conditions"`
	Start      *string            `json:"start,omitempty"`
	End        *string            `json:"end,omitempty"`
}

type BacktestSymbolResult struct {
	Symbol  string `json:"symbol"`
	Signal  []bool `json:"signal"`
	Entries []int  `json:"entries"`
}
type BacktestResp struct {
	BaseTF  domain.Timeframe       `json:"base_timeframe"`
	Results []BacktestSymbolResult `json:"results"`
}

func BacktestRunHandler(reg *engine.Registry, dp engine.DataProvider) fiber.Handler {
	parser := &engine.Parser{Reg: reg}

	return func(c *fiber.Ctx) error {
		var req BacktestReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "bad json"})
		}
		if _, ok := domain.AllowedTF[req.BaseTF]; !ok {
			return c.Status(400).JSON(fiber.Map{"error": "invalid base timeframe"})
		}
		if len(req.Conditions) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "no conditions"})
		}

		// Parse the first condition’s tokens into a predicate
		pred, err := parser.ParsePredicate(req.Conditions[0].Tokens)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		results := make([]BacktestSymbolResult, 0, len(req.Symbols))
		for _, sym := range req.Symbols {
			ctx := engine.NewEvalCtx(sym, req.BaseTF, dp, reg)
			boolSer, err := ctx.EvalPred(pred)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{"error": err.Error(), "symbol": sym})
			}
			results = append(results, BacktestSymbolResult{
				Symbol:  sym,
				Signal:  boolSer,
				Entries: risingEdges(boolSer),
			})
		}
		return c.JSON(BacktestResp{BaseTF: req.BaseTF, Results: results})
	}
}

func risingEdges(b []bool) []int {
	var idxs []int
	for i := 1; i < len(b); i++ {
		if !b[i-1] && b[i] {
			idxs = append(idxs, i)
		}
	}
	return idxs
}
