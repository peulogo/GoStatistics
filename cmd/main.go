package main

import (
	"log"
	"statmq/rabbitmq"
	"statmq/storage"
)

func main() {
	db, err := storage.ConnectDB()
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	rabbitmq.ConsumeMessages(db)
}
