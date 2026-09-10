package models

type BlingWebhookInvoicePayloadType int32

const (
	Pending BlingWebhookInvoicePayloadType = iota
	Issued
	Cancelled
	Rejected
	Denied
	IssuedDANFE
)

type BlingWebhookInvoicePayloadSituation int32

const (
	Incoming BlingWebhookInvoicePayloadType = iota
	Outgoing
)

type BlingWebhookEventTypes interface {
	BlingWebhookPayloadInvoice
}

type BlingWebhookEvent[T BlingWebhookEventTypes] struct {
	EventId   string `json:"eventId"`
	Event     string `json:"event"`
	CompanyId string `json:"companyId"`
	Data      T      `json:"data"`
}

type BlingWebhookPayloadInvoice struct {
	Id            int32                               `json:"id"`
	Type          BlingWebhookInvoicePayloadType      `json:"tipo"`
	Situation     BlingWebhookInvoicePayloadSituation `json:"situacao"`
	Number        string                              `json:"numero"`
	EmissionDate  string                              `json:"dataEmissao"`
	OperationDate string                              `json:"dataOperacao"`
	Contact       struct {
		Id int32 `json:"id"`
	} `json:"contato"`
	OperationNature struct {
		Id int32 `json:"id"`
	} `json:"naturezaOperacao"`
	Store struct {
		Id int32 `json:"id"`
	} `json:"loja"`
}
