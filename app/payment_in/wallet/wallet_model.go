package payment_wallet

type WalletResponse struct {
	Status            string `json:"status_code"`
	Statusmessages    string `json:"status_message"`
	TransactionId     string `json:"transaction_id"`
	OrderId           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	PaymentType       string `json:"payment_type"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	TransactionTime   string `json:"transaction_time"`
	FinishRedirect    string `json:"finish_redirect_url"`
}
