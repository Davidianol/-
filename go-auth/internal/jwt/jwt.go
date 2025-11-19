package jwt

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var secret string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не найден, читаю из переменных окружения")
	}
	secret = os.Getenv("SECRET_KEY")
	if secret == "" {
		log.Fatal("SECRET_KEY не задана в .env или переменных окружения")
	}
}

func CreateAccessToken(id int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"id":   id,
		"role": role,
		"exp":  time.Now().Add(time.Minute * 15).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func CreateRefreshToken(id int64, verison int64) (string, error) {
	claims := jwt.MapClaims{
		"id":      id,
		"version": verison,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseAccesToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("токен невалиден")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}
	// Проверка срока действия токена
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("токен не содержит корректного поля exp")
	}
	if int64(exp) < time.Now().Unix() {
		return nil, fmt.Errorf("токен просрочен")
	}
	return claims, nil
}

// ParseRefreshToken парсит refresh-токен и проверяет срок его действия.
func ParseRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("токен невалиден")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("не удалось преобразовать claims")
	}
	// Проверка срока действия токена
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("токен не содержит корректного поля exp")
	}
	if int64(exp) < time.Now().Unix() {
		return nil, fmt.Errorf("токен просрочен")
	}
	return claims, nil
}
