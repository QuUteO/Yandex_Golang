package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type User struct {
	Id    uuid.UUID `yaml:"Id"`
	Email string    `yaml:"Email"`
	Name  string    `yaml:"Name"`
}

type Config struct {
	Host     string `yaml:"POSTGRES_HOST" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     uint16 `yaml:"POSTGRES_PORT" env:"POSTGRES_PORT" env-default:"5432"`
	Username string `yaml:"POSTGRES_USER" env:"POSTGRES_USER" env-default:"root"`
	Password string `yaml:"POSTGRES_PASS" env:"POSTGRES_PASS" env-default:"1234"`
	Database string `yaml:"POSTGRES_DB" env:"POSTGRES_DB" env-default:"postgres"`

	MaxCon int32 `yaml:"MAX_CON" env:"MAX_CON" env-default:"10"`
	MinCon int32 `yaml:"MIN_CON" env:"MIN_CON" env-default:"5"`
}

// New подключение к постгресу
func New(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	// создается подключение к постгресу
	conString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_max_conns=%d&pool_min_conns=%d&sslmode=disable",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.MaxCon,
		config.MinCon,
	)

	conn, err := pgxpool.New(ctx, conString)

	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к Postgres: %w", err)
	}

	// Применям миграции
	m, err := migrate.New(
		"file://db/migrations",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate to database: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to migrate to database: %w", err)
	}
	return conn, nil
}

func UserExistsByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (bool, error) {
	if pool == nil {
		return false, fmt.Errorf("connection pool is nil")
	}
	var exists bool
	err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// SaveUsers сохранение пользователя
func SaveUsers(ctx context.Context, pool *pgxpool.Pool, user User) error {
	if pool == nil {
		return fmt.Errorf("connection pool is nil")
	}
	// сохраняем пользователя с ID
	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, name, email) VALUES ($1, $2, $3) ON CONFLICT (email) DO NOTHING`,
		user.Id, user.Name, user.Email,
	)

	if err != nil {
		log.Printf("failed to save users: %v", err)
	}
	return err
}

// GetUserDetailsByID В данном случае мы можем добавить функцию для получения email и name по OwnerID
func GetUserDetailsByID(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (string, string, error) {
	var email, name string

	// SQL-запрос для получения данных пользователя
	err := pool.QueryRow(ctx, `SELECT email, name FROM users WHERE id = $1`, userID).Scan(&email, &name)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при получении данных пользователя: %w", err)
	}
	return email, name, nil
}
