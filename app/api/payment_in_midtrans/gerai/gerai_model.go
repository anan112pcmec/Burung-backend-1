package payment_in_gerai

type GeraiResponse struct {
	Status            string `json:"status_code"`
	Statusmessages    string `json:"status_message"`
	TransactionId     string `json:"transaction_id"`
	OrderId           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	PaymentCode       string `json:"payment_code"`
	PdfUrl            string `json:"pdf_url"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	TransactionTime   string `json:"transaction_time"`
	FinishRedirect    string `json:"finish_redirect_url"`
}
