package payment_out_disbursment

type ValidasiResponse interface {
	Validation() bool
}

func (Response401Error) Validation() bool {
	return false
}

func (Response422Error) Validation() bool {
	return false
}

func (ResponseBankAccInquiry) Validation() bool {
	return true
}

func (ResponseDisbursment) Validation() bool {
	return true
}
