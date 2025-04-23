package sendEmail

import (
	"fmt"
	"github.com/google/uuid"
	"net/smtp"
)

func SendEmail(Token, toEmail, toName, fromEmail, appPassword, smtpHost, smtpPort string) error {
	// Подготовка данных
	auth := smtp.PlainAuth("", fromEmail, appPassword, smtpHost)

	// Тема и тело письма
	subject := "Subject: Привет от 06:Go!\r\n"
	body := fmt.Sprintf("Здравствуйте, %s! Добро пожаловать в систему.\r\n Ваша ссылка для регистрации http://localhost:3000/verify/%s\r\n", toName, Token)

	// Финальное сообщение
	msg := []byte(subject + "\r\n" + body)

	// Отправка
	to := []string{toEmail}
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, to, msg)
	if err != nil {
		return fmt.Errorf("не удалось отправить письмо: %w", err)
	}
	return nil
}

func SendEmailToJoin(canvasID, toEmail, toName, fromEmail, appPassword, smtpHost, smtpPort string, OwnerID uuid.UUID) error {
	// Подготовка данных
	auth := smtp.PlainAuth("", fromEmail, appPassword, smtpHost)

	// Тема и тело письма
	subject := "Subject: Команда 06:Go!\r\n"
	body := fmt.Sprintf("%s Ваша ссылка для подключения к canvas_room \r\n http://localhost:3000/canvases/%s/add/%s\r\n", toName, canvasID, OwnerID.String())

	// Финальное сообщение
	msg := []byte(subject + "\r\n" + body)

	// Отправка
	to := []string{toEmail}
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, to, msg)
	if err != nil {
		return fmt.Errorf("не удалось отправить письмо: %w", err)
	}
	return nil
}
