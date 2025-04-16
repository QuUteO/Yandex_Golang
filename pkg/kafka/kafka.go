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
	"time"
)

// User создаем две структуры
// одна с данными о пользователе
// вторая как заглушка

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type Consumer struct {
	Cfg  *config.Config // конфиг приложения (для email)
	Pool *pgxpool.Pool  // пул подключения к базе PostgreSQL
}

// Cleanup CleanUp функиция, которая работает при завершении работы программы
func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer запустился, а Cleanup завершил свою работу")
	return nil
}

// Setup функция, которая работает при запуске программы
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer запустился, а Setup завершил свою работу")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()

	for message := range claim.Messages() {
		// Распаковка JSON-сообщения
		var msg User
		err := json.Unmarshal(message.Value, &msg)
		if err != nil {
			log.Printf("Ошибка при распаковке сообщения: %s\n", err)
			continue
		}

		log.Printf("Обрабатываю пользователя: %+v\n", msg.Name)

		switch message.Topic {
		// если пользователь не существует, то сохраняем в PostgreSQL и отправляем сообщение на почту
		case "register":
			if exist, _ := postgres.UserExists(ctx, c.Pool, msg.Email); exist {
				log.Printf("Пользователь уже был сохранен %s\n", msg.Email)
			} else {
				if err := postgres.SaveUsers(ctx, c.Pool, postgres.User{
					Name:  msg.Name,
					Email: msg.Email,
				}); err != nil {
					return fmt.Errorf("Ошибка сохранения пользователя %w\n", err)
				}
				if err := sendEmail.SendEmail(
					msg.Email,
					msg.Name,
					c.Cfg.Email,
					c.Cfg.Password,
					c.Cfg.Smtphost,
					c.Cfg.Smtpport,
				); err != nil {
					log.Printf("Ошибка отправки сообщения %s\n", err)
				}

			}
		// 	обновляем данные в PostgreSQL и отправляем сообщение на почту
		case "update":
			if err := postgres.UpdateUser(ctx, c.Pool, postgres.User{
				Name:  msg.Name,
				Email: msg.Email,
			}); err != nil {
				return fmt.Errorf("Ошибка сохранения пользователя %s\n", err)
			}
			if err := sendEmail.SendEmail(
				msg.Email,
				msg.Name,
				c.Cfg.Email,
				c.Cfg.Password,
				c.Cfg.Smtphost,
				c.Cfg.Smtpport,
			); err != nil {
				log.Printf("Ошибка отправки сообщения %s\n", err)
			}
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func subscribe(ctx context.Context, topic string, consumerGroup sarama.ConsumerGroup) error {
	consumer := Consumer{}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, &consumer); err != nil {
				log.Printf("Error from consumerGroup consume: %s\n", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}

func StartConsumer(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool) error {

	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	// создаем ConsumerGroup
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.KafkaBrokers, "notification_group", saramaCfg)
	if err != nil {
		return err
	}

	// читаем из топики
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, cfg.Kafka.KafkaTopic, &Consumer{
				Cfg:  cfg,
				Pool: pool,
			}); err != nil {
				log.Printf("Error from consumerGroup consume: %s\n", err)
				time.Sleep(time.Second)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}
