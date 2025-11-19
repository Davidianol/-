package hash

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GetHashPassword(password string) (string, error) {
	// Убираем символ новой строки
	password = strings.TrimSpace(password)
	// Создаем хэш пароля и проверяем его
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return "", fmt.Errorf("ошибка во время генерации пароль: %v. Попробуйте снова", err)

	}
	return string(hashPassword), nil

}
func CompareHashPassword(passwordClient string, passwordServerHash string) bool {
	passwordClient = strings.TrimSpace(passwordClient)
	passwordServerHash = strings.TrimSpace(passwordServerHash)
	err := bcrypt.CompareHashAndPassword([]byte(passwordServerHash), []byte(passwordClient))
	return err == nil
}
