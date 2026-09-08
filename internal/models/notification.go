package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type Topic string

var (
	ALL              Topic = "all"
	PRIVATE          Topic = "private"
	INVOICE_ISSUED   Topic = "invoice-issued"
	RECEIPT_RECEIVED Topic = "receipt-received"
	EMPTY            Topic = ""
)

var ErrInvalidTopic = fmt.Errorf("Invalid topic does not exist")

func GetAllNotificationTopics() []Topic {
	return []Topic{ALL, PRIVATE, INVOICE_ISSUED, RECEIPT_RECEIVED}
}

func GetAllNotificationTopicsWithEmpty() []Topic {
	return append(GetAllNotificationTopics(), EMPTY)
}

func (t *Topic) UnmarshalJSON(data []byte) error {
	var unmarshalled string
	if err := json.Unmarshal(data, &unmarshalled); err != nil {
		return err
	}

	exists := false
	for _, topic := range GetAllNotificationTopics() {
		if Topic(unmarshalled) == topic {
			exists = true
			break
		}
	}

	if !exists {
		return ErrInvalidTopic
	}
	*t = Topic(unmarshalled)
	return nil
}

func (t Topic) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

type Notification struct {
	ID        string            `json:"id"`
	ClientID  string            `json:"client_id"`
	Topic     Topic             `json:"topic"`
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
