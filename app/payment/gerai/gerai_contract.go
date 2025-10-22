package payment_gerai

import (
	"strconv"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

type Response interface {
	Pembayaran() (models.Pembayaran, bool)
}

func Bayar(r Response) (models.Pembayaran, bool) {
	return r.Pembayaran()
}

func (b *GeraiResponse) Pembayaran() (p models.Pembayaran, s bool) {
	s = true

	if b.PaymentType != "cstore" || b.OrderId == "" || b.FraudStatus == "" {
		s = false
	}

	grossFloat, err := strconv.ParseFloat(b.GrossAmount, 64)
	if err != nil {
		s = false
	}

	p = models.Pembayaran{
		KodeTransaksi:      b.TransactionId,
		KodeOrderTransaksi: b.OrderId,
		Provider:           b.PaymentType,
		Amount:             int32(grossFloat),
		PaymentType:        b.PaymentType,
		PaidAt:             b.TransactionTime,
	}
	return
}
