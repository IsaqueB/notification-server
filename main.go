package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/IsaqueB/notification-server/cmd/api"
	"github.com/IsaqueB/notification-server/cmd/worker"
	"github.com/IsaqueB/notification-server/internal/auth"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	runFlag := flag.String("r", "req", "select service to run [api]|[worker] default:[worker]")
	flag.Parse()

	switch *runFlag {
	case "api":
		api.Run()
	case "worker":
		worker.Run()
	case "req":
		aaa()
	default:
		fmt.Println("Invalid service")
	}
}

func aaa() {
	// event := models.BlingWebhookEvent[models.BlingWebhookPayloadInvoice]{
	// 	EventId:   "0",
	// 	Event:     "invoice.issued",
	// 	CompanyId: "0",
	// 	Data: models.BlingWebhookPayloadInvoice{
	// 		Id:            0,
	// 		Type:          models.Issued,
	// 		Situation:     models.Incoming,
	// 		Number:        "0",
	// 		EmissionDate:  "2024-09-27 11:24:56",
	// 		OperationDate: "2024-09-27 11:24:56",
	// 		Contact: models.BlingWebhookPayloadInvoiceRelatedField{
	// 			Id: 0,
	// 		},
	// 		OperationNature: models.BlingWebhookPayloadInvoiceRelatedField{
	// 			Id: 0,
	// 		},
	// 		Store: models.BlingWebhookPayloadInvoiceRelatedField{
	// 			Id: 0,
	// 		},
	// 	},
	// }
	a := `{
		"eventId": "01945027-150e-72b4-e7cf-4943a042cd9c",
		"date": "2025-01-10T12:18:46Z",
		"version": "v1",
		"event": "invoice.updated",
		"companyId": "d4475854366a36c86a37e792f9634a51",
		"data": {
			"id": 12345678,
			"tipo": 1,
			"situacao": 1,
			"numero": "1234",
			"dataEmissao": "2024-09-27 11:24:56",
			"dataOperacao": "2024-09-27 11:00:00",
			"contato": {
				"id": 12345678
			},
			"naturezaOperacao": {
				"id": 12345678
			},
			"loja": {
				"id": 12345678
			}
		}
	}`
	event := &models.BlingWebhookEvent[models.BlingWebhookPayloadInvoice]{}
	err := json.Unmarshal([]byte(a), event)
	if err != nil {
		fmt.Println(err)
	}

	message, err := json.Marshal(event)
	if err != nil {
		fmt.Println(err)
	}

	event2 := &models.BlingWebhookEvent[models.BlingWebhookPayloadInvoice]{}
	err = json.Unmarshal(message, event2)
	if err != nil {
		fmt.Println(err)
	}

	godotenv.Load()
	secret := os.Getenv("BLING_CLIENT_SECRET")
	secretB, _ := hex.DecodeString(secret)
	if err != nil {
		fmt.Println(err)
	}
	signature := hex.EncodeToString(auth.Sign_HS256(message, secretB))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"https://app-notify.bracomil.com.br/api/webhook/bling/invoice",
		bytes.NewBuffer(message),
	)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bling-Signature-256", signature)
	res, err := client.Do(req)
	b, _ := io.ReadAll(res.Body)
	fmt.Println(b)
}
