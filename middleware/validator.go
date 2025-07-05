package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func ValidateRawData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		parts := strings.Fields(string(c.Body()))
		if len(parts) != 3 {
			return c.Status(400).JSON(fiber.Map{
				"error": "expected format: <date> <humidity> <temperature>",
			})
		}

		date, err := time.Parse("2006-01-02", parts[0])

		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid date format"})
		}

		var humidity float64
		var temperature float64

		if _, err := fmt.Sscanf(parts[1], "%f", &humidity); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid humidity value",
			})
		}

		if _, err := fmt.Sscanf(parts[2], "%f", &temperature); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid temperature value",
			})
		}

		fmt.Print(date, humidity, temperature)

		c.Locals("date", date)
		c.Locals("humidity", humidity)
		c.Locals("temperature", temperature)

		return c.Next()
	}

}

func ValidateDate(dateParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		dateString := c.Params(dateParam)
		date, err := time.Parse("2006-01-02", dateString)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid %s format. Use YYYY-MM-DD", dateParam),
			})
		}
		c.Locals("date", date)
		return c.Next()
	}
}

func ValidateDateRange() fiber.Handler {
	return func(c *fiber.Ctx) error {

		startParam, endParam := c.Query("start"), c.Query("end")

		startDate, err1 := time.Parse("2006-01-02", startParam)
		if err1 != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid start date format. Use YYYY-MM-DD"})
		}

		endDate, err2 := time.Parse("2006-01-02", endParam)
		if err2 != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid end date format. Use YYYY-MM-DD"})
		}

		if startDate.After(endDate) {
			return c.Status(400).JSON(fiber.Map{"error": "Start date must be before end date"})
		}

		c.Locals("startDate", startDate)
		c.Locals("endDate", endDate)

		return c.Next()
	}
}
