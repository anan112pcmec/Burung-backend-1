package payment_response_models

type VaNumbers struct {
	Bank     string `json:"bank"`
	VaNumber string `json:"va_number"`
}

type VirtualAccountResponse struct {
	FinishRedirectUrl string      `json:"finish_redirect_url"`
	FraudStatus       string      `json:"fraud_status"`
	GrossAmount       string      `json:"gross_amount"`
	OrderId           string      `json:"order_id"`
	PaymentType       string      `json:"payment_type"`
	PdfUrl            string      `json:"pdf_url"`
	StatusCode        string      `json:"status_code"`
	StatusMessage     string      `json:"status_message"`
	TransactionId     string      `json:"transaction_id"`
	TransactionStatus string      `json:"transaction_status"`
	TransactionTime   string      `json:"transaction_time"`
	VaNumbers         []VaNumbers `json:"va_numbers"`
}

type BcaVirtualAccountResponse struct {
	BcaVaNumber string `json:"bca_va_number"`
	VirtualAccountResponse
}
