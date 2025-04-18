package main

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"project/pkg/api/config"
	"project/pkg/api/logger"
	"project/pkg/api/postgres"
	"project/pkg/kafka"
)

func main() {
	ctx := context.Background() // создаем контекст

	// gracefully shutdown

	// создаем канал вместимостью 1, ипользуя сигналы системы
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	// init logger: zaplogger +
	ctx, _ = logger.New(ctx) // создаем логгер

	// init config: cleanenv +
	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Чтение конфигураций", zap.Error(err))
	}

	//init storage: Postges и migrations +
	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Ошибка подключения к БД", zap.Error(err))
	}

	// run consumer
	err = kafka.StartConsumer(ctx, pool, cfg)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Error Consumer main.go", zap.Error(err))
	}

	select {
	case <-ctx.Done():
		pool.Close()
		stop()
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Server stopped")
	}

}
