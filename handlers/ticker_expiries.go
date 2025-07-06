// handlers/ticker_expiries.go

package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"
)

func GetTickerExpiries() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tickerSymbol := c.Query("ticker")
		dateParam := c.Query("date")
		strikePriceParam := c.Query("strike_price")
		optionTypeParam := c.Query("option_type")

		if tickerSymbol == "" || dateParam == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "ticker and date query parameters are required",
			})
		}

		// Parse date
		fromDate, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid date format, use YYYY-MM-DD",
			})
		}

		// Find ticker ID
		var ticker models.Ticker
		if err := database.DB.Where("ticker_symbol = ?", tickerSymbol).First(&ticker).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "ticker not found",
			})
		}

		// Build query
		query := database.DB.Where("ticker_id = ? AND expiry_date > ?", ticker.ID, fromDate)

		if strikePriceParam != "" {
			strikePrice, err := strconv.ParseFloat(strikePriceParam, 64)
			if err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": "invalid strike_price format",
				})
			}
			query = query.Where("strike_price = ?", strikePrice)
		}

		if optionTypeParam != "" {
			query = query.Where("option_type = ?", optionTypeParam)
		}

		// Execute query
		var expiries []models.TickerExpiry
		if err := query.Find(&expiries).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to fetch ticker expiries",
			})
		}

		// Prepare minimal response
		var result []fiber.Map
		for _, expiry := range expiries {
			result = append(result, fiber.Map{
				"expiry_date":  expiry.ExpiryDate,
				"option_type":  expiry.OptionType,
				"strike_price": expiry.StrikePrice,
				"lot_size":     expiry.LotSize,
			})
		}

		return c.JSON(result)
	}
}
