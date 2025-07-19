package handlers

import (
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"

	"github.com/gofiber/fiber/v2"
)

func GetNewsList(c *fiber.Ctx) error {
	search := c.Query("search") // Optional query param

	var newsList []models.NewsArticle

	query := database.DB.Model(&models.ETNewsArticle{}).Select("title, published_at").Order("published_at DESC").Limit(100)

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	if err := query.Scan(&newsList).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch articles"})
	}

	return c.JSON(newsList)
}
