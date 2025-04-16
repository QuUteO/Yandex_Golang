package sendEmail

import (
	"fmt"
	"net/smtp"
)

func SendEmail(toEmail, toName, fromEmail, appPassword, smtpHost, smtpPort string) error {
	// Подготовка данных
	auth := smtp.PlainAuth("", fromEmail, appPassword, smtpHost)

	// Тема и тело письма
	subject := "Subject: Привет от Go!\r\n"
	body := fmt.Sprintf("Здравствуйте, %s! Добро пожаловать в систему.\r\n", toName)

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
