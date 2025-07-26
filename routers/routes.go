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

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.SendString("Hello over HTTPS!")
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	api := app.Group("/api/v1")

	api.Get("/tickers", handlers.GetTickers)
	api.Get("/expiries", handlers.GetTickerExpiries())
	api.Get("/option_chain", handlers.FetchOptionChain)
	api.Get("/news", handlers.GetNewsList)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// app.Get("/ws", websocket.New(handlers.OhlcWebSocketHandler))
	app.Get("/ws", websocket.New(handlers.WsHandler))
}
