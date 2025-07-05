package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path} [${ip}:${port}] ${time} | Latency: ${latency}\n",
		TimeFormat: "02-Jan-2006 15:04:05", // DD-Mon-YYYY HH:MM:SS
		Output:     os.Stdout,
	})
}
