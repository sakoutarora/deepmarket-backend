package routers

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gulll/deepmarket/handlers"
	"github.com/gulll/deepmarket/middleware"
)

func Setup(app *fiber.App) {
	app.Use(middleware.Logger())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	api := app.Group("/api/v1")

	api.Get("/tickers", handlers.GetTickers)
	api.Get("/expiries", handlers.GetTickerExpiries())
	api.Get("/option_chain", handlers.FetchOptionChain)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(handlers.OhlcWebSocketHandler))
}
