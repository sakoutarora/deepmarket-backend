package handlers

import (
	"log"

	"github.com/gulll/deepmarket/database"

	"github.com/gofiber/fiber/v2"
)

func GetTickers(c *fiber.Ctx) error {
	var symbols []string

	// Only pluck "ticker_symbol" column
	if err := database.DB.Model(&struct {
	}{}).Table("tickers").Pluck("ticker_symbol", &symbols).Error; err != nil {
		log.Printf("Error fetching ticker symbols: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch tickers"})
	}

	return c.JSON(symbols)
}
