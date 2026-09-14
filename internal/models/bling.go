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
	Incoming BlingWebhookInvoicePayloadSituation = iota
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

type BlingWebhookPayloadInvoiceRelatedField struct {
	Id int32 `json:"id"`
}

type BlingWebhookPayloadInvoice struct {
	Id              int32                                  `json:"id"`
	Type            BlingWebhookInvoicePayloadType         `json:"tipo"`
	Situation       BlingWebhookInvoicePayloadSituation    `json:"situacao"`
	Number          string                                 `json:"numero"`
	EmissionDate    string                                 `json:"dataEmissao"`
	OperationDate   string                                 `json:"dataOperacao"`
	Contact         BlingWebhookPayloadInvoiceRelatedField `json:"contato"`
	OperationNature BlingWebhookPayloadInvoiceRelatedField `json:"naturezaOperacao"`
	Store           BlingWebhookPayloadInvoiceRelatedField `json:"loja"`
}

type SheetsWebhookReceipt struct {
	Id             string `json:"id"`
	Party          string `json:"party"`
	Category       string `json:"category"`
	DocumentNumber string `json:"documentNumber"`
	PaymentMethod  string `json:"paymentMethod"`
	User           string `json:"user"`
}
