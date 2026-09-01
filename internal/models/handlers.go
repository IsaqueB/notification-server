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

type WS_ClientMessageType int

const (
	ACK = iota + 1
	HEARTBEAT
)

type WS_ClientMessage struct {
	Type     WS_ClientMessageType `json:"type"`
	ClientId string               `json:"client_id"`
	Title    string               `json:"title"`
	Message  string               `json:"message"`
}

func (m WS_ClientMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":      map[WS_ClientMessageType]string{ACK: "ACK", HEARTBEAT: "HEARTBEAT"}[m.Type],
		"client_id": m.ClientId,
		"title":     m.Title,
		"message":   m.Message,
	})
}

func (m *WS_ClientMessage) UnmarshalJSON(data []byte) error {
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
				m.Type = WS_ClientMessageType(num)
			}
		}
	case float64:
		m.Type = WS_ClientMessageType(v)
	}

	return nil
}
