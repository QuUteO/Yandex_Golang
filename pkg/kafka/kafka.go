package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"project/pkg/api/config"
	"project/pkg/api/postgres"
)

// User создаем две структуры
// одна с данными о пользователе
// вторая как заглушка

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type Consumer struct {
	Cfg  *config.Config
	Pool *pgxpool.Pool
}

// Cleanup CleanUp функиция, которая работает при завершении работы программы
func (c *Consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer Cleanup completed")
	return nil
}

// Setup функция, которая работает при запуске программы
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Printf("Consumer Setup completed")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var msg User
		err := json.Unmarshal(message.Value, &msg)
		if err != nil {
			log.Printf("Error unmarshalling messega: %s\n", err)
		}

		log.Printf("Message claimed: %+v\n", msg.Name)

		switch message.Topic {
		case "register":
			if exist, _ := postgres.UserExists(context.Background(), c.Pool, msg.Email); exist {
				log.Printf("Пользователь уже был сохранен %s\n", msg.Email)
			} else {
				if err := postgres.SaveUsers(context.Background(), c.Pool, postgres.User{
					Name:  msg.Name,
					Email: msg.Email,
				}); err != nil {
					log.Printf("Ошибка сохранения пользователя %s\n", err)
				}
				//from := c.Cfg.Email
				//password := c.Cfg.Password
				//smtpH := c.Cfg.Smtphost
				//smtpP := c.Cfg.Smtpport
				//to := []string{msg.Email}
				//soob :=

			}
		case "update":
			if err := postgres.UpdateUser(context.Background(), c.Pool, postgres.User{
				Name:  msg.Name,
				Email: msg.Email,
			}); err != nil {
				log.Printf("Ошибка обновления данных %s\n", err)
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

//func StartConsumer(ctx context.Context, cfg *config.Config) error {
//
//	saramaCfg := sarama.NewConfig()
//
//	saramaCfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
//	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
//	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetNewest
//
//	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.KafkaBrokers, cfg.Kafka.KafkaTopic, saramaCfg)
//	if err != nil {
//		return err
//	}
//
//	_ = subscribe(ctx, "orders", consumerGroup)
//	time.Sleep(time.Second)
//
//	return subscribe(ctx, "orders", consumerGroup)
//}
