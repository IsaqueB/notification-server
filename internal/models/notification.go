package models

import "time"

type Notification struct {
	ID        string            `json:"id"`
	ClientID  string            `json:"client_id"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Priority  string            `json:"priority"`
	Sound     bool              `json:"sound"`
	Duration  int               `json:"duration"`
	Icon      string            `json:"icon"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
}

type NotificationResponse struct {
	Success        bool      `json:"success"`
	Message        string    `json:"message"`
	Timestamp      time.Time `json:"timestamp"`
	ClientID       string    `json:"client_id"`
	NotificationID string    `json:"notification_id,omitempty"`
}
