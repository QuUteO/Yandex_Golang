package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"project/pkg/api/config"
	"project/pkg/api/postgres"
	"project/pkg/sendEmail"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User структура для пользователя
type User struct {
	UserID uuid.UUID `json:"UserId"`
	Name   string    `json:"Name"`
	Email  string    `json:"Email"`
	Token  string    `json:"Token"`
}
type AddToWhiteListMessage struct {
	CanvasID   string
	CanvasName string // для красоты
	OwnerID    uuid.UUID
	UserId     uuid.UUID // id пользователя, которого добавляют
	Email      string
	Name       string
}

// Consumer структура для consumer
type Consumer struct {
	Cfg  *config.Config // конфиг приложения (для email)
	Pool *pgxpool.Pool  // пул подключения к базе PostgreSQL
}

// Setup функция, которая работает при запуске программы
func (c *Consumer) Setup(sess sarama.ConsumerGroupSession) error {
	log.Println("[Setup] Starting session... Assigned partitions:")
	for topic, partitions := range sess.Claims() {
		log.Printf("  Topic: %s | Partitions: %v\n", topic, partitions)
	}
	return nil
}

// Cleanup CleanUp функиция, которая работает при завершении работы программы
func (c *Consumer) Cleanup(sess sarama.ConsumerGroupSession) error {
	log.Println("[Cleanup] Cleaning up session...")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("[ConsumeClaim] Started for topic %s, partition %d\n", claim.Topic(), claim.Partition())
	ctx := session.Context()

	for message := range claim.Messages() {
		log.Printf("[ConsumerClaim] Получено сообщение: %s ", string(message.Value))

		if len(message.Topic) == 0 {
			log.Printf("Пустой Json")
		}

		switch message.Topic {
		// если пользователь не существует, то сохраняем в PostgreSQL и отправляем сообщение на почту
		case "register":
			// Распаковка JSON-сообщения
			var msg User                               // инициализируем структуру
			err := json.Unmarshal(message.Value, &msg) // парсим сообщение и сохраняем в стурктуру
			log.Printf("Полученные данные %s", string(message.Value))
			if err != nil {
				log.Printf("Ошибка при распаковке сообщения: %s\n", err)
				continue
			}
			log.Printf("Проверка в базе данных...")
			exist, err := postgres.UserExistsByID(ctx, c.Pool, msg.UserID)
			if err != nil {
				log.Printf("Ошибка при проверке существования пользователя: %s\n", err)
				continue
			}
			if exist {
				log.Printf("Пользователь уже был сохранен %s\n", msg.Email)
			} else {
				log.Printf("Попытка сохранения пользователя...")
				if err := postgres.SaveUsers(ctx, c.Pool, postgres.User{
					Id:    msg.UserID,
					Name:  msg.Name,
					Email: msg.Email,
				}); err != nil {
					log.Printf("Ошибка сохранения пользователя: %s\n", err)
					continue
				}
				log.Printf("Попытка отправки сообщения")
				if err := sendEmail.SendEmail(
					msg.Token,
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
		case "joinToCanvas":
			var msg AddToWhiteListMessage
			log.Printf("Проверка в базе данных... для OwnerID: %s", msg.OwnerID)
			_, err := postgres.UserExistsByID(ctx, c.Pool, msg.OwnerID)
			if err != nil {
				log.Printf("Ошибка при проверке существования пользователя: %s\n", err)
				continue
			}
			// Получаем email и имя владельца канваса
			email, name, err := postgres.GetUserDetailsByID(ctx, c.Pool, msg.UserId)
			if err != nil {
				log.Printf("Ошибка при получении данных пользователя: %s\n", err)
				continue
			}
			// Присваиваем email и имя в структуру сообщения
			msg.Email = email
			msg.Name = name

			if err := sendEmail.SendEmailToJoin(
				msg.CanvasID,
				msg.Email,
				msg.Name,
				c.Cfg.Email,
				c.Cfg.Password,
				c.Cfg.Smtphost,
				c.Cfg.Smtpport,
				msg.OwnerID,
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
func subscribe(ctx context.Context, topic string, consumerGroup sarama.ConsumerGroup, pool *pgxpool.Pool, cfg *config.Config) error {
	consumer := Consumer{
		Pool: pool,
		Cfg:  cfg,
	}

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

func StartConsumer(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {

	saramaCfg := sarama.NewConfig()

	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	// создаем ConsumerGroup
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, "test", saramaCfg)
	if err != nil {
		log.Printf("Ошибка при создании consumer group: %s\n", err)
		return err
	}

	return subscribe(ctx, "register", consumerGroup, pool, cfg)
}
