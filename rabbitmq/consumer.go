package rabbitmq

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"statmq/storage"
	"time"

	"github.com/streadway/amqp"
)

func GetClickLogs(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query(`SELECT id, short_url_id, user_agent, ip_address, created_at FROM click_log ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, "Ошибка при получении данных", http.StatusInternalServerError)
		log.Printf("Ошибка запроса в базу данных: %v", err)
		return
	}
	defer rows.Close()

	var logs []storage.ClickLog
	for rows.Next() {
		var logEntry storage.ClickLog
		if err := rows.Scan(&logEntry.ID, &logEntry.ShortURLID, &logEntry.UserAgent, &logEntry.IPAddress, &logEntry.Timestamp); err != nil {
			log.Printf("Ошибка чтения данных из базы: %v", err)
			http.Error(w, "Ошибка при чтении данных", http.StatusInternalServerError)
			return
		}
		logs = append(logs, logEntry)
	}

	response := map[string]interface{}{
		"data": logs,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Ошибка при отправке данных", http.StatusInternalServerError)
		log.Printf("Ошибка при кодировании JSON: %v", err)
		return
	}
}

func StartServer(db *sql.DB) {
	http.HandleFunc("/clicks/log", func(w http.ResponseWriter, r *http.Request) {
		GetClickLogs(w, r, db)
	})

	log.Println("Сервер запущен на порту 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

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

		logEntry.Timestamp = time.Now()

		err := storage.SaveClickLog(db, logEntry)
		if err != nil {
			log.Printf("Ошибка сохранения в БД: %v", err)
		} else {
			log.Printf("Сохранено  %s", logEntry.Timestamp)
		}
	}
}
