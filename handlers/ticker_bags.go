package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/models"
)

func GetTickerBags(c *fiber.Ctx) error {
	// Local type for a "bag" (collection of tickers)
	type BagRow struct {
		ID      int                     `json:"id"`
		Name    string                  `json:"name"`
		Symbols []models.TickerResponse `json:"symbols"`
	}

	// Example data (replace with DB query later)
	bags := []BagRow{
		{
			ID:   1,
			Name: "Nifty 50",
			Symbols: []models.TickerResponse{
				{ID: 1, TradingSymbol: "SBIN"},
				{ID: 2, TradingSymbol: "RELIANCE"},
				{ID: 3, TradingSymbol: "HDFCBANK"},
				{ID: 4, TradingSymbol: "TATAMOTORS"},
			},
		},
	}

	resp := models.APIResponse{
		Success: true,
		Message: "Ticker bags fetched successfully",
		Data:    bags,
	}
	return c.JSON(resp)
}
