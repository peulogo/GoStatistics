package storage

import "time"

type ClickLog struct {
	ID         int       `json:"id"`
	ShortURLID int       `json:"short_url_id"`
	UserAgent  string    `json:"user_agent"`
	IPAddress  string    `json:"ip_address"`
	Timestamp  time.Time `json:"created_at"`
}
