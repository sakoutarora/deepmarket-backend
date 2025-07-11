package routers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"github.com/gulll/deepmarket/handlers"
	"github.com/gulll/deepmarket/middleware"
)

func Setup(app *fiber.App) {
	app.Use(middleware.Logger())
	app.Get("/tickers", handlers.GetTickers)
	app.Get("/expiries", handlers.GetTickerExpiries())
	app.Get("api/v1/option_chain", handlers.FetchOptionChain)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(handlers.OhlcWebSocketHandler))
}
