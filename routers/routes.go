package routers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"github.com/gulll/deepmarket/handlers"
)

func Setup(app *fiber.App) {
	app.Get("/tickers", handlers.GetTickers)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(handlers.OhlcWebSocketHandler))
}
