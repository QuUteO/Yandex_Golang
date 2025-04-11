package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	Host     string `yaml:"POSTGRES_HOST" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     uint16 `yaml:"POSTGRES_PORT" env:"POSTGRES_PORT" env-default:"5432"`
	Username string `yaml:"POSTGRES_USER" env:"POSTGRES_USER" env-default:"root"`
	Password string `yaml:"POSTGRES_PASS" env:"POSTGRES_PASS" env-default:"1234"`
	Database string `yaml:"POSTGRES_DB" env:"POSTGRES_DB" env-default:"postgres"`

	MaxCon int32 `yaml:"MAX_CON" env:"MAX_CON" env-default:"10"`
	MinCon int32 `yaml:"MIN_CON" env:"MIN_CON" env-default:"5"`
}

// подключение к постгресу
func New(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	// создается подключение к постгресу
	conString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s&pool_max_conns=%d&pool_min_conns=%d",
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

	return conn, nil
}
