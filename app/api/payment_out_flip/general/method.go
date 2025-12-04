package payment_out_general

func (r ResponseFlipGeneralWrapper) Validating() bool {
	if r.Err401 != nil || r.Err422 != nil {
		return false
	}

	return true
}

func (r ResponseFlipGeneralWrapper) ReturnGetBalance() ResponseGetBalance {
	return *r.GetBalance
}

func (r ResponseFlipGeneralWrapper) ReturnGetBank() []ResponseGetBank {
	return *r.GetBank
}
