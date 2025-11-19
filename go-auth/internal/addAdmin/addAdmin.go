package main

import (
	"Auth/internal/db"
	"Auth/internal/hash"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

func readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Добавляем перевод строки после ввода пароля
	if err != nil {
		return "", err
	}
	return string(passwordBytes), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Инициализируем соединение с БД
	err := db.InitDB(ctx)
	if err != nil {
		fmt.Printf("Ошибка подключения к базе данных: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Утилита добавления пользователя ===")

	// Считываем данные от пользователя
	login, err := readInput("Логин: ")
	if err != nil {
		fmt.Printf("Ошибка ввода: %v\n", err)
		os.Exit(1)
	}

	password, err := readPassword("Пароль: ")
	if err != nil {
		fmt.Printf("Ошибка ввода пароля: %v\n", err)
		os.Exit(1)
	}

	role, err := readInput("Роль (ADMIN/CLIENT): ")
	if err != nil {
		fmt.Printf("Ошибка ввода: %v\n", err)
		os.Exit(1)
	}

	// Приводим роль к верхнему регистру и проверяем на валидность
	role = strings.ToUpper(role)
	if role != "ADMIN" && role != "CLIENT" {
		fmt.Println("Неверная роль. Допустимые значения: ADMIN, CLIENT.")
		os.Exit(1)
	}

	// Хешируем пароль
	hashPassword, err := hash.GetHashPassword(password)
	if err != nil {
		fmt.Printf("Ошибка хеширования пароля: %v\n", err)
		os.Exit(1)
	}

	// Создаем пользователя
	user := db.User{
		Login:    login,
		Password: hashPassword,
		Role:     role,
	}

	// Добавляем пользователя в БД
	userOut, err := db.AddUser(ctx, user)
	if err != nil {
		fmt.Printf("Ошибка добавления пользователя: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nПользователь успешно добавлен!\n")
	fmt.Printf("ID: %d\n", userOut.ID)
	fmt.Printf("Логин: %s\n", userOut.Login)
	fmt.Printf("Роль: %s\n", userOut.Role)
}
