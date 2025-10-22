package payment_wallet

import (
	"fmt"
	"strconv"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

type Response interface {
	Pembayaran() (models.Pembayaran, bool)
}

func Bayar(r Response) (models.Pembayaran, bool) {
	return r.Pembayaran()
}

func (b *WalletResponse) Pembayaran() (models.Pembayaran, bool) {
	m := models.Pembayaran{}
	var s bool = true

	fmt.Println("[TRACE] Mulai proses Pembayaran WalletResponse")
	fmt.Printf("[TRACE] Data masuk: OrderId=%s, TransactionId=%s, PaymentType=%s, GrossAmount=%s\n",
		b.OrderId, b.TransactionId, b.PaymentType, b.GrossAmount)

	if b.OrderId == "" || b.TransactionId == "" || b.PaymentType != "qris" {
		fmt.Println("[TRACE] Data tidak valid: salah satu field kosong atau PaymentType bukan qris")
		s = false
		return m, s
	}

	grossFloat, err := strconv.ParseFloat(b.GrossAmount, 64)
	if err != nil {
		fmt.Printf("[TRACE] Error konversi GrossAmount (%s): %v\n", b.GrossAmount, err)
		s = false
		return m, s
	}

	fmt.Printf("[TRACE] GrossAmount berhasil dikonversi ke float64: %.2f\n", grossFloat)
	grossInt := int32(grossFloat)
	fmt.Printf("[TRACE] GrossAmount dikonversi ke int32: %d\n", grossInt)

	fmt.Println("[TRACE] Membuat objek models.Pembayaran...")

	m = models.Pembayaran{
		KodeTransaksi:      b.TransactionId,
		KodeOrderTransaksi: b.OrderId,
		Provider:           "wallet",
		Amount:             grossInt,
		PaymentType:        b.PaymentType,
		PaidAt:             b.TransactionTime,
	}

	fmt.Printf("[TRACE] Pembayaran selesai dibuat: %+v\n", m)
	fmt.Println("[TRACE] Selesai proses Pembayaran WalletResponse")

	return m, s
}
