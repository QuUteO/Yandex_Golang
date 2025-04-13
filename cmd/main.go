package main

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/smtp"
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
	pool, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Ошибка подключения к БД", zap.Error(err))
	}

	// todo: init router: chi
	r := chi.NewRouter()
	r.Post("/onRegister", SaveUserHandler(cfg, pool))
	r.Put("/onSend", SaveUserHandler(cfg, pool))

	// todo: run server
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Запуск сервера на http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Ошибка запуска сервера: ", zap.Error(err))
	}
}

// SaveUsers сохранение пользователя
func SaveUsers(ctx context.Context, pool *pgxpool.Pool, user postgres.User) error {
	// пишем сохранение users
	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, name, email) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = $2, email = $3`,
		user.Id, user.Name, user.Email,
	)

	if err != nil {
		log.Printf("failed to save users: %v", err)
	}
	return err
}

// UpdateUser пишем обновление users
func UpdateUser(ctx context.Context, pool *pgxpool.Pool, user postgres.User) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET name = $1, email = $2 WHERE id = $3`,
		user.Name, user.Email, user.Id,
	)

	if err != nil {
		log.Printf("failed to save users: %v", err)
	}
	return err
}

func SaveUserHandler(cfg *config.Config, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверка метода запроса
		var user postgres.User
		// Декодируем тело запроса
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Printf("Ошибка декодирования JSON: ", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// используем сохранение пользователя
		err = SaveUsers(r.Context(), pool, user)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// отправка сообщения на почту
		smtpHost := cfg.Smtphost
		smtpPort := cfg.Smtpport
		from := cfg.Email
		appPassword := cfg.Password
		// получатель
		to := []string{user.Email}

		// Сообщение
		subject := "Subject: Привет от Go!\r\n"
		body := "Это тестовое письмо, отправленное из Go с использованием Gmail App Password.\r\n"
		msg := []byte(subject + "\r\n" + body)

		// Аутентификация
		auth := smtp.PlainAuth("", from, appPassword, smtpHost)

		if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("Отправка прошла успешно")

		// отправляем ответ
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte("Пользователь сохранён и уведомлен по email"))
	}
}

func UpdateUserHandler(cfg *config.Config, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// проверка на метод
		if r.Method == http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var user postgres.User

		// декодирование сообщения
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// обновление данных БД
		err = UpdateUser(r.Context(), pool, user)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// отправка сообщения на почту
		smtpHost := cfg.Smtphost
		smtpPort := cfg.Smtpport
		from := cfg.Email
		appPassword := cfg.Password
		// получатель
		to := []string{user.Email}

		// Сообщение
		subject := "Subject: Привет от Go!\r\n"
		body := "Это тестовое письмо, отправленное из Go с использованием Gmail App Password.\r\n"
		msg := []byte(subject + "\r\n" + body)

		// Аутентификация
		auth := smtp.PlainAuth("", from, appPassword, smtpHost)

		if err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("Отправка прошла успешно")

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte("Пользователь обновлён и уведомлён по email"))
	}
}
