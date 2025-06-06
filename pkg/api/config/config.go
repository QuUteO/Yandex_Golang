package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"project/pkg/api/postgres"
	"time"
)

type Config struct {
	Postgres     postgres.Config `yaml:"POSTGRES" env:"POSTGRES"`
	HTTPServer   `yaml:"HTTPSERVER"`
	Notification `yaml:"NOTIFICATION"`
	Kafka        KafkaConfig `yaml:"KAFKA"`
}

type HTTPServer struct {
	Address     string        `yaml:"ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"IDLE_TIMEOUT" env-default:"60s"`
}

type Notification struct {
	Smtphost string `yaml:"SMTP_HOST" env-default:"smtp.gmail.com"`
	Smtpport string `yaml:"SMTP_PORT" env-default:"587"`
	Email    string `yaml:"EMAIL" env-default:"test@gmail.com"`
	Password string `yaml:"PASSWORD" env-default:"test"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env-default:"kafka1:19092,kafka2:19093,kafka3:19094"`
	Topics  []string `yaml:"topics"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig("./config/config.yaml", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
