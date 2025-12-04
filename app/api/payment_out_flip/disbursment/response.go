package payment_out_disbursment

type Response401Error struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int32  `json:"status"`
}

type Response422Error struct {
	Code   string `json:"code"`
	Errors []struct {
		Attribute string `json:"attribute"`
		Code      int64  `json:"code"`
		Message   string `json:"message"`
	} `json:"errors"`
}

type ResponseBankAccInquiry struct {
	BankCode         string `json:"bank_code"`
	AccountNumber    string `json:"account_number"`
	AccountHolder    string `json:"account_holder"`
	Status           string `json:"status"`
	InquiryKey       string `json:"inquiry_key"`
	IsVirtualAccount bool   `json:"is_virtual_account"`
}

type ResponseDisbursment struct {
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
	RecipientCity    int64  `json:"recipient_city"`
	CreatedFrom      string `json:"created_from"`
	Direction        string `json:"direction"`
	Sender           string `json:"sender"` // "null" dikirim sebagai string, bukan null
	Fee              int64  `json:"fee"`
	BeneficiaryEmail string `json:"beneficiary_email"`
	IdempotencyKey   string `json:"idempotency_key"`
	IsVirtualAccount bool   `json:"is_virtual_account"`
}

type ResponseDisbursmentWrapper struct {
	Error401               *Response401Error
	Error422               *Response422Error
	ResponseBankAccInq     *ResponseBankAccInquiry
	ResponseDisbursment    *ResponseDisbursment
	ResponseAllDisbursment *[]ResponseDisbursment
}
