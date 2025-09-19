package pengguna_transaction_services

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type PayloadCheckoutBarangCentang struct {
	IDPengguna   int64              `json:"id_pengguna_checkout_barang"`
	Username     string             `json:"username_pengguna_checkout_barang"`
	DataCheckout []models.Keranjang `json:"data_checkout"`
}
