package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла: ", err)
	}
}

func ConnectDB() (*sql.DB, error) {

	log.Printf("Connecting to DB with user: %s, host: %s, port: %s",
		os.Getenv("DB_USER"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}

func SaveClickLog(db *sql.DB, logEntry ClickLog) error {
	_, err := db.Exec(`
		INSERT INTO click_logs (timestamp, page_url, short_url, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5)`,
		logEntry.Timestamp,
		logEntry.PageURL,
		logEntry.ShortURL,
		logEntry.UserAgent,
		logEntry.IPAddress,
	)
	return err
}
