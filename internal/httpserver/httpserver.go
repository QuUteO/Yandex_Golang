package httpserver

import (
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"net/smtp"
	"project/pkg/api/postgres"
	"project/pkg/config"
)

func SaveUserHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверка метода запроса
		if r.Method == http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var user config.Users
		// Декодируем тело запроса
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// используем сохранение пользователя
		err = postgres.SaveUsers(r.Context(), pool, &user)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// отаправка сообщения на почту
		if err := sendMail(); err != nil {
			http.Error(w, "Error send message", http.StatusInternalServerError)
			return
		}

		// отправляем ответ
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte("Пользователь сохранён и уведомлен по email"))
	}
}

func UpdateUserHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// проверка на метод
		if r.Method == http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var user config.Users

		// декодирование сообщения
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// обновление данных БД
		err = postgres.UpdateUser(r.Context(), pool, &user)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// отправка сообщения
		if err := sendMail(); err != nil {
			http.Error(w, "Error send message", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte("Пользователь обновлён и уведомлён по email"))
	}
}

func sendMail() error {
	// информация об отправители
	from := "quuteo86@gmail.com"
	password := "13tengr13"

	// информация о получателе
	to := []string{
		"ignatievmisha039@gmail.com",
	}

	// smtp сервер конфигураций
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// сообщение
	message := []byte("Тестовой сообщение через golang.")

	//авторизация
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Отправка почты.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)

	return err
}
