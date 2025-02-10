package rabbitmq

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"statmq/storage"

	"github.com/streadway/amqp"
)

func ConsumeMessages(db *sql.DB) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Ошибка подключения к RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Ошибка открытия канала: %v", err)
	}
	defer ch.Close()

	exchangeName := "statistic_service"
	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошибка создания exchange: %v", err)
	}

	queueName := "statistic_service.click_log"
	queue, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошибка создания очереди: %v", err)
	}

	routingKey := "click.log"
	err = ch.QueueBind(
		queue.Name,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошибка привязки очереди: %v", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Ошибка при подписке на очередь: %v", err)
	}

	log.Println("Listening for messages...")

	for msg := range msgs {
		var logEntry storage.ClickLog
		if err := json.Unmarshal(msg.Body, &logEntry); err != nil {
			log.Printf("Ошибка парсинга сообщения: %v", err)
			continue
		}

		err := storage.SaveClickLog(db, logEntry)
		if err != nil {
			log.Printf("Ошибка сохранения в БД: %v", err)
		} else {
			log.Printf("Сохранено %s", logEntry.Timestamp)
		}
	}
}
