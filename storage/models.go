package storage

type ClickLog struct {
	Timestamp string `json:"timestamp"`
	PageURL   string `json:"page_url"`
	ShortURL  string `json:"short_url"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}
