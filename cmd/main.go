package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/routers"
)

// func main() {

// 	if err := config.SetupConfig(); err != nil {
// 		log.Fatalf("config SetupConfig() error: %s", err)
// 	}
// 	masterDSN := config.DbConfiguration()

// 	if err := database.DbConnection(masterDSN); err != nil {
// 		log.Fatalf("database DbConnection error: %s", err)
// 	}
// 	server := routers.SetupRoute()
// 	port := viper.GetString("SERVER_PORT")
// 	log.Fatalf("%v", server.Run(":"+port))
// }

func main() {
	database.Init()

	app := fiber.New()

	routers.Setup(app)

	log.Println("Server running on :8080")
	log.Fatal(app.Listen(":8080"))
}
