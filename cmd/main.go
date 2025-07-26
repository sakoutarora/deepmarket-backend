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

	log.Println("Server running on :8080")
	err := app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
