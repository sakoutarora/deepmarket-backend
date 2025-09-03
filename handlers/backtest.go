package handlers

import (
	"github.com/gulll/deepmarket/backtesting/adapters"
	"github.com/gulll/deepmarket/backtesting/controller"
	domain "github.com/gulll/deepmarket/backtesting/domain"
	engine "github.com/gulll/deepmarket/backtesting/engine"
	"github.com/gulll/deepmarket/models"

	"github.com/gofiber/fiber/v2"
)

func BacktestRunHandler(reg *engine.Registry, dp engine.DataProvider) fiber.Handler {
	parser := &engine.Parser{Reg: reg}

	return func(c *fiber.Ctx) error {
		var req domain.BacktestReq
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(models.APIResponse{
				Success: false,
				Message: "Invalid Request format " + err.Error(),
			})
		}

		if _, ok := domain.AllowedTF[req.BaseTF]; !ok {
			return c.Status(400).JSON(models.APIResponse{
				Success: false,
				Message: "Invalid Base Timeframe",
			})
		}

		// --- ENTRY PLAN ---
		entryPred, err := parser.ParsePredicate(req.EntryConditions.Tokens)
		if err != nil {
			return c.Status(400).JSON(models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}

		planner := engine.NewPlanner(req.BaseTF)
		entryPlan, err := planner.Build(entryPred)
		if err != nil {
			return c.Status(400).JSON(models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}

		// --- EXIT PLAN (optional) ---
		var exitPlan *engine.Plan // default nil
		if req.ExitConditions != nil && len(req.ExitConditions.Tokens) > 0 {
			exitPred, err := parser.ParsePredicate(req.ExitConditions.Tokens)
			if err != nil {
				return c.Status(400).JSON(models.APIResponse{
					Success: false,
					Message: err.Error(),
				})
			}

			planner = engine.NewPlanner(req.BaseTF)
			exitPlan, err = planner.Build(exitPred)
			if err != nil {
				return c.Status(400).JSON(models.APIResponse{
					Success: false,
					Message: err.Error(),
				})
			}
		}

		// --- DATA LOADING ---
		ctx := engine.NewEvalCtx(req.Symbol, req.BaseTF, dp, reg)
		ohlc, err := dp.LoadOHLCV(req.Symbol, req.BaseTF)
		if err != nil {
			return c.Status(500).JSON(models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}

		ctx.SetCache(adapters.CandlesToSeries(ohlc))
		rt := engine.NewRuntime(ctx)

		// --- RUN BACKTEST ---
		trades, _, equity, err := controller.RunBacktest(
			req, req.Symbol, ctx, rt, entryPlan, exitPlan, ohlc,
		)
		if err != nil {
			return c.Status(500).JSON(models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
		}

		// --- SUMMARY ---
		summary := controller.ComputeSummary(trades, equity, float64(req.Capital))

		return c.JSON(models.APIResponse{
			Success: true,
			Message: "Backtest completed",
			Data:    summary,
		})
	}
}
