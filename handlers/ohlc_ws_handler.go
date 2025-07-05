package handlers

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"
)

func OhlcWebSocketHandler(c *websocket.Conn) {
	symbol := c.Query("symbol")
	if symbol == "" {
		log.Println("Missing symbol parameter")
		_ = c.WriteMessage(websocket.TextMessage, []byte("Missing symbol parameter"))
		c.Close()
		return
	}

	log.Printf("WebSocket client connected for symbol: %s", symbol)

	for {
		data, err := fetchLatestOHLC(symbol)
		if err != nil {
			log.Printf("Fetch error: %v", err)
			break
		}

		if err := c.WriteJSON(data); err != nil {
			log.Printf("Write error: %v", err)
			break
		}

		time.Sleep(5 * time.Second)
	}
}

func fetchLatestOHLC(symbol string) ([]models.OptionNiftyOHLC, error) {
	var data []models.OptionNiftyOHLC

	err := database.DB.Where("symbol = ?", symbol).
		Order("candle_time desc").
		Limit(10).
		Find(&data).Error

	return data, err
}
