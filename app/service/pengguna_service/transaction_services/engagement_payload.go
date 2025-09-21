package pengguna_transaction_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"

)

type PayloadCheckoutBarangCentang struct {
	IDPengguna   int64              `json:"id_pengguna_checkout_barang"`
	Username     string             `json:"username_pengguna_checkout_barang"`
	DataCheckout []models.Keranjang `json:"data_checkout"`
}

type PayloadSnapTransaksiRequest struct {
	UserInformation   models.Pengguna                                    `json:"data_user_transaksi"`
	AlamatInformation models.AlamatPengguna                              `json:"data_alamat_transaksi"`
	DataCheckout      response_transaction_pengguna.ResponseDataCheckout `json:"data_items_transaksi"`
}
