package routers

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	engine "github.com/gulll/deepmarket/backtesting/engine"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/handlers"
	"github.com/gulll/deepmarket/middleware"
	"github.com/gulll/deepmarket/models"
)

func Setup(app *fiber.App) {
	app.Use(middleware.Logger())

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(models.APIResponse{
			Success: true,
			Message: "API is healthy",
			Data:    nil,
		})
	})

	e := engine.BuildRegistry()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	api := app.Group("/api/v1")

	api.Get("/tickers", handlers.GetTickers)
	api.Get("/ticker/bags", handlers.GetTickerBags)
	api.Get("/expiries", handlers.GetTickerExpiries())
	api.Get("/option_chain", handlers.FetchOptionChain)
	api.Post("/condition/validate", handlers.ValidateConditionHandler(e))
	api.Post("/backtest", handlers.BacktestRunHandler(e, engine.NewPGProvider(database.DB)))

	app.Get("/news", handlers.GetNewsList)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Google Client ID:", os.Getenv("GOOGLE_CLIENT_ID"))

	app.Post("/api/login", handlers.LoginHandler())
	app.Post("/api/signup", handlers.SignupHandler())
	app.Get("/api/auth/google", handlers.GoogleOAuthHandler())
	app.Get("/api/auth/google/callback", handlers.GoogleCallbackHandler())
	app.Get("/ws", websocket.New(handlers.WsHandler))
}
