package handlers

import (
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/utils/options"
)

type OptionDetail struct {
	OI         string `json:"oi,omitempty"`
	Volume     string `json:"volume,omitempty"`     // Placeholder, calculate if you want
	VolPercent string `json:"volPercent,omitempty"` // Placeholder, calculate if you want
	IV         string `json:"iv,omitempty"`         // Placeholder, calculate if you want
	LTP        string `json:"ltp,omitempty"`        // open price
	LTPChg     string `json:"ltpChg,omitempty"`     // Placeholder, calculate if you want
}

type OptionData struct {
	StrikePrice string        `json:"strikePrice"`
	Calls       *OptionDetail `json:"calls,omitempty"`
	Puts        *OptionDetail `json:"puts,omitempty"`
}

func FetchOptionChain(c *fiber.Ctx) error {

	// Querry Param Ticker Expiry and current time
	ticker := c.Query("ticker")
	expiry := c.Query("expiry")
	currentTime := c.Query("current_time")

	if ticker == "" || expiry == "" || currentTime == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ticker, expiry and current_time query parameters are required",
		})
	}

	// Query Database for option chain
	var optionChain []struct {
		StrikePrice float64 `json:"strike_price"`
		OptionType  string  `json:"option_type"`
		OI          *int64  `json:"oi,omitempty"`
		Open        float64 `json:"open"`
	}

	var strike float64

	err := database.DB.Table("ohlc_data_nse_eq").
		Select("open").
		Where("ticker = ? and time = ?", strings.ToLower(ticker), "2025-06-11 09:15:00").
		Scan(&strike).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Spot price not found",
		})
	}

	dbTable := "option_stock_ohlc"

	if ticker == "NIFTY" || ticker == "BANKNIFTY" {
		dbTable = "option_nifty_ohlc"
	}

	err = database.DB.Table(dbTable).
		Select("strike_price, option_type, oi, open").
		Where("symbol = ? AND expiry_date = ? AND candle_time = ?", ticker, expiry, currentTime).
		Scan(&optionChain).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch option chain",
		})
	}

	// Calculate expiry - currentTime in days
	expiryDate, err := time.Parse("2006-01-02", expiry)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid expiry date format"})
	}
	currentDate, err := time.Parse("2006-01-02 15:04:05", currentTime)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid current_time format"})
	}
	daysToExpiry := expiryDate.Sub(currentDate).Hours() / 24

	type OptionSide struct {
		LTP   float64 `json:"ltp"`
		OI    *int64  `json:"oi"`
		IV    float64 `json:"iv"`
		DELTA float64 `json:"delta"`
		VEGA  float64 `json:"vega"`
		THETA float64 `json:"theta"`
	}
	type OptionRow struct {
		StrikePrice string      `json:"strikePrice"`
		Calls       *OptionSide `json:"calls,omitempty"`
		Puts        *OptionSide `json:"puts,omitempty"`
	}

	strikeMap := map[float64]*OptionRow{}

	for _, row := range optionChain {
		entry, exists := strikeMap[row.StrikePrice]
		if !exists {
			entry = &OptionRow{
				StrikePrice: formatNumber(row.StrikePrice),
			}
			strikeMap[row.StrikePrice] = entry
		}

		side := &OptionSide{
			LTP: row.Open,
			OI:  row.OI,
		}

		if row.OptionType == "CE" {

			optionType := options.Call
			S := strike               // spot price
			K := row.StrikePrice      // strike
			T := daysToExpiry / 365.0 // time to expiry in years
			r := 0.065                // risk-free rate
			ltp := side.LTP           // observed option price

			iv := options.ImpliedVolatility(ltp, S, K, T, r, optionType)
			side.DELTA = options.Delta(S, K, T, r, iv, optionType)
			side.VEGA = options.Vega(S, K, T, r, iv)
			side.THETA = options.Theta(S, K, T, r, iv, optionType)

			side.IV = iv
			entry.Calls = side

		} else if row.OptionType == "PE" {

			optionType := options.Put
			S := strike
			K := row.StrikePrice
			T := daysToExpiry / 365.0
			r := 0.065
			ltp := side.LTP

			iv := options.ImpliedVolatility(ltp, S, K, T, r, optionType)
			side.DELTA = options.Delta(S, K, T, r, iv, optionType)
			side.VEGA = options.Vega(S, K, T, r, iv)
			side.THETA = options.Theta(S, K, T, r, iv, optionType)

			entry.Puts = side
		}
	}

	// Convert map to slice
	var finalOptions []OptionRow
	for _, val := range strikeMap {
		finalOptions = append(finalOptions, *val)
	}

	return c.JSON(finalOptions)
}

func formatNumber(f float64) string {
	return humanize.CommafWithDigits(f, 2)
}
