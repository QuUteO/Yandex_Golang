package main

import (
	"context"
	"go.uber.org/zap"
	"log"
	"project/internal/config"
	"project/pkg/api/logger"
	"project/pkg/api/postgres"
)

func main() {
	ctx := context.Background() // создаем контекст

	// todo: init logger: zaplogger +
	ctx, _ = logger.New(ctx) // создаем логгер

	// todo: init config: cleanenv +
	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Чтение конфигураций", zap.Error(err))
	} else {
		log.Println("Все успешно завелось конфиг")
	}

	//todo: init storage: Postges +
	_, err = postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Ошибка подключения к БД")
	} else {
		log.Println("Успешно завелся постгрес")
	}
	// todo: init router: chi

	// todo: run server
}
