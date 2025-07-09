package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"
)

func OhlcWebSocketHandler(c *websocket.Conn) {
	tickerSymbol := c.Query("ticker")
	expiryParam := c.Query("expiry")
	strikePriceParam := c.Query("strike_price")
	optionTypeParam := c.Query("option_type")

	if tickerSymbol == "" || expiryParam == "" {
		log.Println("ticker and date query parameters are required")
		_ = c.WriteMessage(websocket.TextMessage, []byte("ticker and expiry query parameters are required"))
		c.Close()
		return
	}
	expiry, err := time.Parse("2006-01-02", expiryParam)

	if err != nil {
		_ = c.WriteMessage(websocket.TextMessage, []byte("invalid date format, use YYYY-MM-DD"))
		c.Close()
		return
	}

	log.Printf("WebSocket client connected for symbol: %s", tickerSymbol)

	strikePrice, err1 := strconv.ParseFloat(strikePriceParam, 64)

	if err1 != nil {
		_ = c.WriteMessage(websocket.TextMessage, []byte("invalid strike_price format"))
		c.Close()
		return
	}
	startTimeStr := c.Query("start_time", "")
	var startTime time.Time
	if startTimeStr != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04:05", startTimeStr)
		if err != nil {
			log.Printf("Invalid start_time: %v", err)
			_ = c.WriteMessage(websocket.TextMessage, []byte("Invalid start_time (must be RFC3339)"))
			return
		}
		startTime = parsedTime
	}
	// for {
	// 	data, err := fetchLatestOHLC(tickerSymbol, expiry, strikePrice, optionTypeParam)
	// 	if err != nil {
	// 		log.Printf("Fetch error: %v", err)
	// 		break
	// 	}

	// 	if err := c.WriteJSON(data); err != nil {
	// 		log.Printf("Write error: %v", err)
	// 		break
	// 	}

	// 	time.Sleep(5 * time.Second)
	// }

	streamOHLC(c, tickerSymbol, expiry, strikePrice, optionTypeParam, startTime)
}

// func fetchLatestOHLC(symbol string, expiry time.Time, strike float64, otype string) ([]models.OptionNiftyOHLC, error) {
// 	var data []models.OptionNiftyOHLC

// 	err := database.DB.Where("symbol = ? AND expiry_date = ? AND strike_price = ? AND option_type = ?", symbol, expiry, strike, otype).
// 		Order("candle_time desc").
// 		Limit(10).
// 		Find(&data).Error

// 	return data, err
// }

// func fetchLatestOHLC(symbol string, expiry time.Time, strike float64, otype string, after time.Time) ([]models.OptionNiftyOHLC, error) {
// 	var data []models.OptionNiftyOHLC
// 	err := database.DB.
// 		Where("symbol = ? AND expiry_date = ? AND strike_price = ? AND option_type = ?", symbol, expiry, strike, otype).
// 		Order("candle_time asc").
// 		Limit(10).
// 		Find(&data).Error
// 	return data, err
// }

// Function to fetch candles newer than lastTime
func fetchOHLCAfter(symbol string, expiry time.Time, strike float64, otype string, after time.Time) ([]models.OptionNiftyOHLC, error) {
	var data []models.OptionNiftyOHLC
	err := database.DB.
		Where("symbol = ? AND expiry_date = ? AND strike_price = ? AND option_type = ? AND candle_time > ?", symbol, expiry, strike, otype, after).
		Order("candle_time asc"). // Important: order ascending when getting newer data
		Limit(10).
		Find(&data).Error
	return data, err
}

// WebSocket streaming function
func streamOHLC(conn *websocket.Conn, symbol string, expiry time.Time, strike float64, otype string, startTime time.Time) {
	var lastTime time.Time

	// If startTime provided, use it directly
	if !startTime.IsZero() {
		lastTime = startTime
	}

	// Continuous loop
	for {
		newData, err := fetchOHLCAfter(symbol, expiry, strike, otype, lastTime)
		if err != nil {
			log.Printf("Fetch error: %v", err)
			break
		}

		if len(newData) > 0 {
			lastTime = newData[len(newData)-1].CandleTime

			if err := conn.WriteJSON(newData); err != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}

		time.Sleep(5 * time.Second)
	}

	log.Println("⛔ Streaming stopped")

}
