package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"project/pkg/api/config"
	"project/pkg/api/postgres"
	"project/pkg/sendEmail"
)

// User структура для пользователя
type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Consumer структура для consumer
type Consumer struct {
	Cfg  *config.Config // конфиг приложения (для email)
	Pool *pgxpool.Pool  // пул подключения к базе PostgreSQL
}

// Cleanup CleanUp функиция, которая работает при завершении работы программы
func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Setup функция, которая работает при запуске программы
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()

	for message := range claim.Messages() {
		// Распаковка JSON-сообщения
		var msg User                               // инициализируем структуру
		err := json.Unmarshal(message.Value, &msg) // парсим сообщение и сохраняем в стурктуру
		if err != nil {
			log.Printf("Ошибка при распаковке сообщения: %s\n", err)
			continue
		}

		log.Printf("Имя пользователя и email: %+v %+v\n", msg.Name, msg.Email)

		switch message.Topic {
		// если пользователь не существует, то сохраняем в PostgreSQL и отправляем сообщение на почту
		case "register":
			exist, err := postgres.UserExists(ctx, c.Pool, msg.Email)
			if err != nil {
				log.Printf("Ошибка при проверке существования пользователя: %s\n", err)
				continue
			}
			if exist {
				log.Printf("Пользователь уже был сохранен %s\n", msg.Email)
			} else {
				if err := postgres.SaveUsers(ctx, c.Pool, postgres.User{
					Name:  msg.Name,
					Email: msg.Email,
				}); err != nil {
					log.Printf("Ошибка сохранения пользователя: %s\n", err)
					continue
				}
				if err := sendEmail.SendEmail(
					msg.Email,
					msg.Name,
					c.Cfg.Email,
					c.Cfg.Password,
					c.Cfg.Smtphost,
					c.Cfg.Smtpport,
				); err != nil {
					log.Printf("Ошибка отправки сообщения на email: %s\n", err)
				}
			}
		// обновляем данные в PostgreSQL и отправляем сообщение на почту
		case "update":
			if err := postgres.UpdateUser(ctx, c.Pool, postgres.User{
				Name:  msg.Name,
				Email: msg.Email,
			}); err != nil {
				log.Printf("Ошибка обновления пользователя: %s\n", err)
				continue
			}
			if err := sendEmail.SendEmail(
				msg.Email,
				msg.Name,
				c.Cfg.Email,
				c.Cfg.Password,
				c.Cfg.Smtphost,
				c.Cfg.Smtpport,
			); err != nil {
				log.Printf("Ошибка отправки сообщения на email: %s\n", err)
			}
		}

		// Отметка о том, что сообщение было успешно обработано
		session.MarkMessage(message, "Message processed successfully")
	}

	return nil
}

// чтение топиков
func subscribe(ctx context.Context, topic string, consumerGroup sarama.ConsumerGroup) error {
	consumer := Consumer{}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, &consumer); err != nil {
				fmt.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

var brokers = []string{"127.0.0.1:9095", "127.0.0.1:9096", "127.0.0.1:9097"}

func StartConsumer(ctx context.Context) error {

	saramaCfg := sarama.NewConfig()

	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	// создаем ConsumerGroup
	consumerGroup, err := sarama.NewConsumerGroup(brokers, "analitic", saramaCfg)

	if err != nil {
		log.Printf("Ошибка при создании consumer group: %s\n", err)
		return err
	}

	return subscribe(ctx, "user", consumerGroup)
}
