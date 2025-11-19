package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	// "Auth/internal/jwt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

var dbx *sqlx.DB

type User struct {
	ID       int64  `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
	Role     string `db:"role"`
	Phone    string `db:"phone"`
	Mail     string `db:"mail"`
}

func GetFromEnv(item string) string {
	result := os.Getenv(item)
	if result == "" {
		log.Fatalf("%s не задана в .env или переменных окружения\n", item)
	}
	return result
}

func InitDB(ctx context.Context) error {
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не найден, читаю из переменных окружения")
	}
	go_user := GetFromEnv("DB_USER")
	go_password := GetFromEnv("DB_PASSWORD")
	go_address := GetFromEnv("DB_ADDRESS")
	go_port := GetFromEnv("DB_PORT")
	go_db_name := GetFromEnv("DB_NAME")

	line := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", go_user, go_password, go_address, go_port, go_db_name)
	for i := 0; i < 10; i++ {
		dbx, err = sqlx.ConnectContext(ctx, "pgx", line)
		if err == nil {
			if pingErr := dbx.PingContext(ctx); pingErr == nil {
				log.Println("Успешно подключились к базе данных!")
				query := `
				CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				login TEXT UNIQUE NOT NULL,
				password TEXT NOT NULL,
				role  TEXT NOT NULL,
				phone  VARCHAR(20) CHECK (phone ~ '^\+?[0-9]+$'),
				mail VARCHAR(255) UNIQUE CHECK (mail ~ '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$'),
				token_version INTEGER NOT NULL DEFAULT 0
				);`
				_, err = dbx.ExecContext(ctx, query)
				if err != nil {
					return fmt.Errorf("ошибка создания таблицы: %w", err)
				}
				return nil
			}
		}
		log.Printf("Попытка подключения #%d неудачна. Повтор через 2 секунды...", i+1)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("не удалось подключиться к БД после нескольких попыток: %w", err)
}
func AddUser(ctx context.Context, user User) (User, error) {
	var userOut User
	var id int64
	err := dbx.GetContext(ctx, &id, `SELECT id FROM users WHERE login=$1`, user.Login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			query := `
			INSERT INTO users (login, password, role) VALUES ($1, $2, $3)`
			_, err := dbx.ExecContext(ctx, query, user.Login, user.Password, user.Role)
			if err != nil {
				return userOut, fmt.Errorf("ошибка вставки пользователя: %w", err)
			}
			err = dbx.GetContext(ctx, &userOut, `
				SELECT id, login, role,
				COALESCE(mail, '') as mail,
				COALESCE(phone, '') as phone
				FROM users
				WHERE login=$1`, user.Login)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return userOut, fmt.Errorf("пользователь не появился в базе данных: %w", err)
				}
				return userOut, fmt.Errorf("Ошибка при проверке пользователя: %w", err)
			}
			return userOut, nil
		} else {
			return userOut, fmt.Errorf("Ошибка при проверке пользователя: %w", err)
		}
	}
	return userOut, fmt.Errorf("ошибка создания пользователя, он уже есть")
}

func GetUserDataById(ctx context.Context, id int64) (User, error) {
	var out User
	err := dbx.GetContext(ctx, &out, `SELECT id, login, role, password FROM users WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, fmt.Errorf("Пользователь с ID = %d не существует", id)
		}
		return out, fmt.Errorf("ошибка получения базовых данных пользователя: %w", err)
	}
	return out, nil

}
func UpdateTokenById(ctx context.Context, id int64) (int64, error) {
	query := `UPDATE users
	SET token_version = token_version + 1
	WHERE id = $1
	RETURNING token_version;`
	var newVersion int64
	err := dbx.GetContext(ctx, &newVersion, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("Пользователь с ID = %d не существует", id)
		}
		return 0, fmt.Errorf("ошибка получения данных пользователя: %w", err)
	}
	return newVersion, nil
}

// getTokenVersion возвращает текущее значение token_version для пользователя с указанным id.
func GetTokenVersion(ctx context.Context, id int64) (int64, error) {
	var version int64
	query := `SELECT token_version FROM users WHERE id = $1`
	err := dbx.GetContext(ctx, &version, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("пользователь с ID = %d не существует", id)
		}
		return 0, fmt.Errorf("ошибка получения token_version: %w", err)
	}
	return version, nil
}
func GetPasswordByLogin(ctx context.Context, login string) (string, error) {
	var password string
	err := dbx.GetContext(ctx, &password, `
		SELECT password FROM users
		WHERE login=$1`, login)
	if err != nil {
		return "", fmt.Errorf("ошибка получения пароля пользователя: %w", err)
	}
	return password, nil
}

func GetProfileUserDataById(ctx context.Context, id int64) (User, error) {
	var out User
	err := dbx.GetContext(ctx, &out, `
		SELECT id, login, role,
		COALESCE(mail, '') as mail,
		COALESCE(phone, '') as phone
		FROM users
		WHERE id=$1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, fmt.Errorf("Пользователь с ID = %d не существует", id)
		}
		return out, fmt.Errorf("ошибка получения профильных данных пользователя: %w", err)
	}
	return out, nil

}

func GetProfileUserDataByLogin(ctx context.Context, login string) (User, error) {
	var out User
	err := dbx.GetContext(ctx, &out, `
		SELECT id, login, role,
		COALESCE(mail, '') as mail,
		COALESCE(phone, '') as phone
		FROM users
		WHERE login=$1`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, fmt.Errorf("%s не существует", login)
		}
		return out, fmt.Errorf("ошибка получения профильных данных пользователя: %w", err)
	}
	return out, nil

}

func GetRoleById(ctx context.Context, id int64) (string, error) {
	var role string
	query := `SELECT role FROM users WHERE id = $1`
	err := dbx.GetContext(ctx, &role, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("пользователь с ID = %d не существует", id)
		}
		return "", fmt.Errorf("ошибка получения роли пользователя: %w", err)
	}
	return role, nil
}

func UpdateProfileUserDataById(ctx context.Context, id int64, login string, mail, phone string) error {
	loginDb := ""
	err := dbx.GetContext(ctx, &loginDb, `SELECT login FROM users WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("ошибка получения данных пользователя: %w", err)
	}
	idDb := int64(0)
	err = dbx.GetContext(ctx, &idDb, `SELECT id FROM users WHERE login=$1`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			query := `
        UPDATE users
        SET login = $1, mail = $2, phone = $3
        WHERE id = $4`
			_, err := dbx.ExecContext(ctx, query, login, mail, phone, id)
			if err != nil {
				return fmt.Errorf("ошибка обновления данных пользователя: %w", err)
			}
			return nil
		}
		return fmt.Errorf("ошибка получения данных пользователя: %w", err)
	} else {
		if id == idDb {
			query := `
        UPDATE users
        SET login = $1, mail = $2, phone = $3
        WHERE id = $4`
			_, err := dbx.ExecContext(ctx, query, login, mail, phone, id)
			if err != nil {
				return fmt.Errorf("ошибка обновления данных пользователя: %w", err)
			}
			return nil
		}
		return fmt.Errorf("пользователь с таким логином уже существует")
	}
}
func UpdatePasswordById(ctx context.Context, id int64, password string) error {
	query := `
        UPDATE users
        SET password = $1 WHERE id = $2`
	_, err := dbx.ExecContext(ctx, query, password, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления пароля пользователя: %w", err)
	}
	return nil
}
