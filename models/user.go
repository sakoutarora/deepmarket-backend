package models

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	Email        string `gorm:"unique;not null"`
	Phone        string `gorm:"unique"`
	PasswordHash string
	FirstName    string
	LastName     string
	Gender       string
	DateOfBirth  *time.Time
	City         string
	State        string
	Country      string
	ProfileURL   string
	Bio          string
	Occupation   string
	Company      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
