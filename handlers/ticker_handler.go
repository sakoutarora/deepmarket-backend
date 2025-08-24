package handlers

import (
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"
)

func GetTickers(c *fiber.Ctx) error {

	// Query params
	q := strings.ToUpper(strings.TrimSpace(c.Query("q", "")))
	limitStr := strings.TrimSpace(c.Query("limit", "20"))

	// Parse & clamp limit
	limit := 20
	if n, err := strconv.Atoi(limitStr); err == nil {
		if n < 1 {
			limit = 1
		} else if n > 100 {
			limit = 100
		} else {
			limit = n
		}
	}

	var tickers []models.TickerResponse

	db := database.DB.
		Table("tickers").
		Select("id, trading_symbol")

	// Prefix match if q provided (case-insensitive)
	if q != "" {
		// Postgres ILIKE for case-insensitive match
		db = db.Where("trading_symbol ILIKE ?", q+"%")
	}

	// Order + Limit
	if err := db.
		Order("trading_symbol ASC").
		Limit(limit).
		Scan(&tickers).Error; err != nil {
		log.Printf("Error fetching ticker symbols: %v", err)
		return c.Status(500).JSON(models.APIResponse{
			Success: false,
			Message: "Failed to fetch tickers",
			Data:    nil,
		})
	}

	return c.JSON(models.APIResponse{
		Success: true,
		Message: "Tickers fetched successfully",
		Data:    tickers,
	})
}
