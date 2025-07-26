package routers

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/gulll/deepmarket/handlers"
)

func Setup(app *fiber.App) {
	app.Get("/tickers", handlers.GetTickers)
	app.Get("/expiries", handlers.GetTickerExpiries())
	app.Get("/news", handlers.GetNewsList)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(handlers.OhlcWebSocketHandler))

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("Google Client ID:", os.Getenv("GOOGLE_CLIENT_ID"))

	app.Post("/api/login", handlers.LoginHandler())
	app.Post("/api/signup", handlers.SignupHandler())
	app.Get("/api/auth/google", handlers.GoogleOAuthHandler())
	app.Get("/api/auth/google/callback", handlers.GoogleCallbackHandler())
}
