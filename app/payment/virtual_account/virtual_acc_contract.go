package payment_va

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

func (b *BcaVirtualAccountResponse) Pembayaran() (p models.Pembayaran, s bool) {
	s = true

	if b.BcaVaNumber == "" || b.OrderId == "" || b.FraudStatus == "" {
		s = false
	}

	grossFloat, err := strconv.ParseFloat(b.GrossAmount, 64)
	if err != nil {
		s = false
	}
	provider := ""
	if len(b.VaNumbers) > 0 {
		provider = b.VaNumbers[0].Bank
	}

	p = models.Pembayaran{
		KodeTransaksi:      b.TransactionId,
		KodeOrderTransaksi: b.OrderId,
		Provider:           provider,
		Amount:             int32(grossFloat),
		PaymentType:        b.PaymentType,
		PaidAt:             b.TransactionTime,
	}
	return
}

// BNI
func (b *BniVirtualAccountResponse) Pembayaran() (p models.Pembayaran, s bool) {
	s = true

	if len(b.VaNumbers) == 0 || b.VaNumbers[0].Bank != "bni" {
		s = false
	}
	if b.OrderId == "" || b.FraudStatus == "" {
		s = false
	}

	grossFloat, err := strconv.ParseFloat(b.GrossAmount, 64)
	if err != nil {
		s = false
	}

	provider := ""
	if len(b.VaNumbers) > 0 {
		provider = b.VaNumbers[0].Bank
	}

	p = models.Pembayaran{
		KodeTransaksi:      b.TransactionId,
		KodeOrderTransaksi: b.OrderId,
		Provider:           provider,
		Amount:             int32(grossFloat),
		PaymentType:        b.PaymentType,
		PaidAt:             b.TransactionTime,
	}
	return
}

func (b *BriVirtualAccountResponse) Pembayaran() (p models.Pembayaran, s bool) {
	s = true

	if len(b.VaNumbers) == 0 || b.VaNumbers[0].Bank != "bri" {
		s = false
	}
	if b.OrderId == "" || b.FraudStatus == "" {
		s = false
	}

	grossFloat, err := strconv.ParseFloat(b.GrossAmount, 64)
	if err != nil {
		s = false
	}

	provider := ""
	if len(b.VaNumbers) > 0 {
		provider = b.VaNumbers[0].Bank
	}

	p = models.Pembayaran{
		KodeTransaksi:      b.TransactionId,
		KodeOrderTransaksi: b.OrderId,
		Provider:           provider,
		Amount:             int32(grossFloat),
		PaymentType:        b.PaymentType,
		PaidAt:             b.TransactionTime,
	}
	return
}

// PERMATA
func (b *PermataVirtualAccount) Pembayaran() (p models.Pembayaran, s bool) {
	s = true

	if b.PermataVaNumber == "" || b.OrderId == "" || b.FraudStatus == "" {
		s = false
	}

	grossFloat, err := strconv.ParseFloat(b.GrossAmount, 64)
	if err != nil {
		s = false
	}

	p = models.Pembayaran{
		KodeTransaksi:      b.TransactionId,
		KodeOrderTransaksi: b.OrderId,
		Provider:           "permata",
		Amount:             int32(grossFloat),
		PaymentType:        b.PaymentType,
		PaidAt:             b.TransactionTime,
	}
	return
}
