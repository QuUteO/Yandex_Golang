package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"project/pkg/api/postgres"
	"time"
)

type Config struct {
	Postgres postgres.Config `yaml:"POSTGRES" env:"POSTGRES"`

	HTTPServer `yaml:"HTTPSERVER"`
}

type HTTPServer struct {
	Address     string        `yaml:"ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"IDLE_TIMEOUT" env-default:"60s"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("./config/config.yaml", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
