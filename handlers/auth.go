package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gulll/deepmarket/database"
	"github.com/gulll/deepmarket/models"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignupHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input SignupInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Could not hash password"})
		}

		user := models.User{
			Username:     input.Username,
			Email:        input.Email,
			PasswordHash: string(hashedPassword),
		}

		if err := database.DB.Create(&user).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Email or username already exists"})
		}

		token, err := generateJWT(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Token generation failed"})
		}

		return c.JSON(fiber.Map{
			"token": token,
			"user":  user.Username,
		})
	}
}

func LoginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input LoginInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input"})
		}

		var user models.User
		if err := database.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "User not found"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Incorrect password"})
		}

		token, err := generateJWT(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Token generation failed"})
		}

		return c.JSON(fiber.Map{
			"token": token,
			"user":  user.Username,
		})
	}
}

func GoogleOAuthHandler() fiber.Handler {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	return func(c *fiber.Ctx) error {
		redirect := c.Query("redirect_uri")
		if redirect == "" {
			redirect = "http://localhost:5173/auth-success"
		}

		state := url.QueryEscape(redirect)
		url := googleOauthConfig.AuthCodeURL(state)
		return c.Redirect(url)
	}
}

func GoogleCallbackHandler() fiber.Handler {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_OAUTH_REDIRECT"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	return func(c *fiber.Ctx) error {
		code := c.Query("code")
		redirectEncoded := c.Query("state")
		redirect, _ := url.QueryUnescape(redirectEncoded)

		log.Println("code and redirect:", redirect)

		token, err := googleOauthConfig.Exchange(context.Background(), code)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Token exchange failed")
		}

		client := googleOauthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to get user info")
		}
		defer resp.Body.Close()

		var userInfo struct {
			ID      string `json:"id"`
			Email   string `json:"email"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to decode user info")
		}

		var user models.User
		if err := database.DB.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
			user = models.User{
				Email:      userInfo.Email,
				Username:   userInfo.Email,
				FirstName:  userInfo.Name,
				ProfileURL: userInfo.Picture,
			}
			if err := database.DB.Create(&user).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
			}
		}

		// Generate JWT
		jwtToken, err := generateJWT(user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Token generation failed")
		}

		// Redirect back to frontend with token
		if redirect != "" {
			return c.Redirect(fmt.Sprintf("%s?token=%s", redirect, jwtToken))
		}

		// Fallback
		return c.JSON(fiber.Map{
			"token": jwtToken,
			"user":  user,
		})
	}
}

func generateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}
