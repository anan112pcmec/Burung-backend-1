package payment_out_disbursment

type PayloadBankAccountInquiry struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	InquiryKey    string `json:"inquiry_key"`
}

type PayloadCreateDisbursment struct {
	AccountNumber    string `json:"account_number"`
	BankCode         string `json:"bank_code"`
	Amount           string `json:"amount"`
	Remark           string `json:"remark"`
	ReciepentCity    string `json:"recipient_city"`
	BeneficiaryEmail string `json:"beneficiary_email"`
}

type PayloadGetDisburstmentById struct {
	Id string `json:"id_disbursment"`
}

type PayloadGetDisburstmentByIdempotencyKey struct {
	IdempotencyKey string `json:"idempotency_key"`
}
