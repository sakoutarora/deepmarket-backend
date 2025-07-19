package models

import (
	"time"
)

type ETNewsArticle struct {
	ID             int       `json:"id"`
	MSID           int       `json:"msid"`
	Title          string    `json:"title"`
	SEOPath        string    `json:"seo_path"`
	ImageCaption   *string   `json:"image_caption"`
	PublishedAt    time.Time `json:"published_at"`
	CreatedAt      time.Time `json:"created_at"`
	ContentFetched bool      `json:"content_fetched"`
}

// Tell GORM explicit table name
func (ETNewsArticle) TableName() string {
	return "et_news_articles"
}

type NewsArticle struct {
	Title       string    `json:"title"`
	PublishedAt time.Time `json:"published_at"`
}
