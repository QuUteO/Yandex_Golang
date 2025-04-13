package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"project/pkg/api/logger"
	"project/pkg/api/postgres"
	"project/pkg/config"
)

func main() {
	ctx := context.Background() // создаем контекст

	// todo: init logger: zaplogger +
	ctx, _ = logger.New(ctx) // создаем логгер

	// todo: init config: cleanenv +
	cfg, err := config.New()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Чтение конфигураций", zap.Error(err))
	}

	//todo: init storage: Postges и migrations +
	_, err = postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Ошибка подключения к БД", zap.Error(err))
	}
	// todo: init router: chi
	router := chi.NewRouter()

	//middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	// todo: run server
}
