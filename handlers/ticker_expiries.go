// handlers/ticker_expiries.go

package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
)

type ExpiryDate struct {
	ExpiryDate time.Time `json:"expiry_date"`
	LotSize    float32   `json:"lot_size"`
}

func GetTickerExpiries() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tickerSymbol := c.Query("ticker")
		dateParam := c.Query("date")

		if tickerSymbol == "" || dateParam == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "ticker and date query parameters are required",
			})
		}

		// Parse the date
		fromDate, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid date format, use YYYY-MM-DD",
			})
		}

		// Raw results
		var rawResults []ExpiryDate

		err = database.DB.
			Table("ticker_expiries").
			Select("ticker_expiries.expiry_date, ticker_expiries.lot_size").
			Joins("LEFT JOIN tickers ON ticker_expiries.ticker_id = tickers.id").
			Where("tickers.ticker_symbol = ? AND ticker_expiries.expiry_date > ?", tickerSymbol, fromDate).
			Group("ticker_expiries.expiry_date, ticker_expiries.lot_size").
			Scan(&rawResults).Error

		if err != nil {
			fmt.Printf("DB error: %v\n", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "failed to fetch distinct expiry dates",
			})
		}

		// Format only date part (YYYY-MM-DD)
		type ExpiryDateResponse struct {
			ExpiryDate string  `json:"expiry_date"`
			LotSize    float32 `json:"lot_size"`
		}

		var response []ExpiryDateResponse
		for _, item := range rawResults {
			response = append(response, ExpiryDateResponse{
				ExpiryDate: item.ExpiryDate.Format("2006-01-02"),
				LotSize:    item.LotSize,
			})
		}

		return c.JSON(response)
	}
}
