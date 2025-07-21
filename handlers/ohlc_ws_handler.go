package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"
)

type OHLC struct {
	Ticker string    `json:"ticker"`
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
	OI     float64   `json:"oi"`
}

type InitOrRangeRequest struct {
	Type      string `json:"type"` // "init" or "range"
	Symbol    string `json:"symbol"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Interval  int    `json:"interval"`
}

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

func WsHandler(c *websocket.Conn) {
	defer c.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		var req InitOrRangeRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		log.Println("Received request:", req)

		start := parseEpoch(req.StartTime)
		end := parseEpoch(req.EndTime)

		rows, err := database.DB.Raw(`
			WITH base AS (
				SELECT
					*,
					ROW_NUMBER() OVER (ORDER BY "time") AS global_row_num
				FROM ohlc_data_nse_eq
				WHERE ticker = ? AND "time" >= ? AND "time" <= ?
			),
			grouped AS (
				SELECT *,
					(global_row_num - 1) / ? AS group_id,
					ROW_NUMBER() OVER (PARTITION BY (global_row_num - 1) / ? ORDER BY "time") AS row_in_group,
					COUNT(*) OVER (PARTITION BY (global_row_num - 1) / ?) AS rows_in_group
				FROM base
			)
			SELECT
				MIN("time") AS interval_start,
				MAX(CASE WHEN row_in_group = 1 THEN open END) AS open,
				MAX(high) AS high,
				MIN(low) AS low,
				MAX(CASE WHEN row_in_group = rows_in_group THEN close END) AS close,
				SUM(volume) AS volume
			FROM grouped
			GROUP BY group_id
			ORDER BY group_id;
		`, req.Symbol, start, end, req.Interval, req.Interval, req.Interval).Rows()

		if err != nil {
			log.Println("DB query failed:", err)
			c.WriteJSON(fiber.Map{"error": "Database error"})
			continue
		}
		defer rows.Close()

		ohlc := make([]OHLC, 0)

		for rows.Next() {
			var ohlci OHLC
			if err := rows.Scan(&ohlci.Time, &ohlci.Open, &ohlci.High, &ohlci.Low, &ohlci.Close, &ohlci.Volume); err != nil {
				log.Println("Scan error:", err)
				continue
			}
			ohlc = append(ohlc, ohlci)
			// if data, err := json.Marshal(ohlc); err == nil {
			// 	c.WriteMessage(websocket.TextMessage, data)
			// }
		}

		if raw, err := json.Marshal(ohlc); err == nil {
			if compressed, err := gzipCompress(raw); err == nil {
				c.WriteMessage(websocket.BinaryMessage, compressed)
			}
		}

	}
}

func parseEpoch(ts int64) time.Time {
	if ts > 1e12 {
		return time.UnixMilli(ts)
	}
	return time.Unix(ts, 0)
}

func gzipCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	gz.Close()
	return buf.Bytes(), nil
}
