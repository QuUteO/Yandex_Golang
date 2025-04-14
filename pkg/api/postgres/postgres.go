package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	Id    int    `yaml:"Id"`
	Email string `yaml:"Email"`
	Name  string `yaml:"name"`
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
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
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
