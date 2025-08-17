package handlers

import (
	"log"

	"github.com/gulll/deepmarket/database"

	"github.com/gofiber/fiber/v2"
)

func GetTickers(c *fiber.Ctx) error {
	type Ticker struct {
		ID             int    `json:"id"`
		TradingSymbol  string `json:"trading_symbol"`
	}

	var tickers []Ticker

	if err := database.DB.
		Table("tickers").
		Select("id, trading_symbol").
		Scan(&tickers).Error; err != nil {
		log.Printf("Error fetching ticker symbols: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tickers"})
	}

	return c.JSON(tickers)
}
