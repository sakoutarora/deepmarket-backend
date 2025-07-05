package token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func GenerateToken(user_id uuid.UUID, user_email string) (string, error) {

	token_lifespan := TOKEN_HOUR_LIFESPAN

	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = user_id
	claims["email"] = user_email
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(token_lifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	JWT_SECRET := viper.GetString("JWT_SECRET")
	return token.SignedString([]byte(JWT_SECRET))

}

func TokenValid(c *gin.Context) error {
	tokenString := ExtractToken(c)
	JWT_SECRET := viper.GetString("JWT_SECRET")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWT_SECRET), nil
	})
	if err != nil {
		return err
	}
	var userID string
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok = claims["user_id"].(string) // Assuming user_id is a string
		if !ok {
			return errors.New("user_id not found in claims or not a string")
		}
	}
	c.Request.Header.Set("user_id", string(userID))
	return nil
}

func ExtractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}
