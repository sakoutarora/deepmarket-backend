package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/routers"
)

func main() {
	database.Init()

	app := fiber.New()

	routers.Setup(app)

	log.Println("Server running on :443")
	err := app.ListenTLS(":443", "./cmd/cert.pem", "./cmd/key.pem") // Use 443 or any port
	if err != nil {
		log.Fatal(err)
	}
}
