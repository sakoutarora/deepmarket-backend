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

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from DeepMarket!",
			"endpoints": fiber.Map{
				"/api/v1/tickers":      "Get a list of all supported tickers",
				"/api/v1/expiries":     "Get a list of all supported expiry dates for a given ticker",
				"/api/v1/option_chain": "Get a list of all supported option strikes for a given ticker and expiry date",
				"/api/v1/news":         "Get a list of all news articles for a given ticker",
				"/ws":                  "Websocket endpoint for fetching OHLC data",
			},
		})
	})

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
