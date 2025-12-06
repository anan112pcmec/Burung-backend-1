package callback_payment_out

type PayloadUpdateStatusPaymentOut struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	Amount           int64  `json:"amount"`
	Status           string `json:"status"`
	Reason           string `json:"reason"`
	Timestamp        string `json:"timestamp"`
	BankCode         string `json:"bank_code"`
	AccountNumber    string `json:"account_number"`
	RecipientName    string `json:"recipient_name"`
	SenderBank       string `json:"sender_bank"`
	Remark           string `json:"remark"`
	Receipt          string `json:"receipt"`
	TimeServed       string `json:"time_served"`
	BundleID         int64  `json:"bundle_id"`
	CompanyID        int64  `json:"company_id"`
	RecipientCity    int    `json:"recipient_city"`
	CreatedFrom      string `json:"created_from"`
	Direction        string `json:"direction"`
	Sender           string `json:"sender"` // "null" dikirim sebagai string, bukan null
	Fee              int    `json:"fee"`
	BeneficiaryEmail string `json:"beneficiary_email"`
	IdempotencyKey   string `json:"idempotency_key"`
	IsVirtualAccount bool   `json:"is_virtual_account"`
}
