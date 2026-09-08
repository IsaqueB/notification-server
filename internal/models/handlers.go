package models

import (
	"encoding/json"
	"strconv"
)

type RequestMessage struct {
	ClientId string `json:"client_id"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type WebsocketClientMessageType int

const (
	ACK = iota + 1
	HEARTBEAT
)

type WebsocketClientMessage struct {
	Type     WebsocketClientMessageType `json:"type"`
	ClientId string                     `json:"client_id"`
	Title    string                     `json:"title"`
	Message  string                     `json:"message"`
}

func (m WebsocketClientMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":      map[WebsocketClientMessageType]string{ACK: "ACK", HEARTBEAT: "HEARTBEAT"}[m.Type],
		"client_id": m.ClientId,
		"title":     m.Title,
		"message":   m.Message,
	})
}

func (m *WebsocketClientMessage) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	m.ClientId, _ = raw["client_id"].(string)
	m.Title, _ = raw["title"].(string)
	m.Message, _ = raw["message"].(string)

	switch v := raw["type"].(type) {
	case string:
		switch v {
		case "ACK":
			m.Type = ACK
		case "HEARTBEAT":
			m.Type = HEARTBEAT
		default:
			if num, err := strconv.Atoi(v); err == nil {
				m.Type = WebsocketClientMessageType(num)
			}
		}
	case float64:
		m.Type = WebsocketClientMessageType(v)
	}

	return nil
}
