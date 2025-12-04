package payment_out_disbursment

func (r ResponseDisbursmentWrapper) Validating() bool {
	if r.Error401 != nil || r.Error422 != nil {
		return false
	}
	return true
}

func (r ResponseDisbursmentWrapper) ReturnResBankInq() ResponseBankAccInquiry {
	return *r.ResponseBankAccInq
}

func (r ResponseDisbursmentWrapper) ReturnDisburstment() ResponseDisbursment {
	return *r.ResponseDisbursment
}
