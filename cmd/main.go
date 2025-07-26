package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/routers"
	"github.com/joho/godotenv"
)

func main() {
	database.Init()

	app := fiber.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	routers.Setup(app)

	log.Println("Server running on :8080")
	log.Fatal(app.Listen(":8080"))
}
