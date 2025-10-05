package pengguna_transaction_services

import (
	"github.com/midtrans/midtrans-go/snap"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/payment/payment_response_models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/identity_pengguna"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services/response_transaction_pengguna"
)

type PayloadCheckoutBarangCentang struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"data_identitas_pengguna"`
	DataCheckout      []models.Keranjang                 `json:"data_checkout"`
	JenisLayananKurir string                             `json:"jenis_layanan_kurir_checkout_barang"`
}

type PayloadSnapTransaksiRequest struct {
	UserInformation   models.Pengguna                                    `json:"data_user_transaksi"`
	AlamatInformation models.AlamatPengguna                              `json:"data_alamat_transaksi"`
	DataCheckout      response_transaction_pengguna.ResponseDataCheckout `json:"data_transaksi_item"`
}

type PayloadReactionTransaksiSnap struct {
	response_transaction_pengguna.SnapTransaksi
	snap.Response
}

type PayloadLockTransaksi struct {
	DataHold      []response_transaction_pengguna.CheckoutData      `json:"checkout_data_hold"`
	PaymentResult payment_response_models.BcaVirtualAccountResponse `json:"payment_result"`
	IdAlamatUser  int64                                             `json:"alamat_data_hold"`
}
