package response_transaction_pengguna

import (
	"github.com/midtrans/midtrans-go/snap"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

type CheckoutData struct {
	IDUser           int64  `json:"id_user_barang_keranjang"`
	IDSeller         int32  `json:"id_seller_barang_keranjang"`
	NamaSeller       string `json:"nama_seller_barang_keranjang"`
	JenisBarang      string `json:"jenis_barang_keranjang"`
	IdBarangInduk    int32  `json:"id_barang_induk_keranjang"`
	IdKategoriBarang int64  `json:"id_kategori_barang_keranjang"`
	NamaBarang       string `json:"nama_barang_keranjang"`
	NamaKategori     string `json:"nama_kategori_barang_keranjang"`
	HargaKategori    int32  `json:"harga_barang_kategori_keranjang"`
	Dipesan          int32  `json:"dipesan_barang_keranjang"`
	Status           bool   `json:"status_barang_keranjang"`
	Message          string `json:"pesan_data_keranjang"`
}

type LayananPengiriman struct {
	JenisLayananKurir string `json:"jenis_layanan_kurir_keranjang"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Response Struct Checkout
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ResponseDataCheckout struct {
	Message           string            `json:"pesan_chekout_barang"`
	DataResponse      []CheckoutData    `json:"data_response_checkout_barang"`
	LayananPengiriman LayananPengiriman `json:"data_layanan_pengiriman_barang"`
}

type ResponseBatalCheckout struct {
	Message string `json:"pesan_batal_checkout"`
}

type ResponseDataValidateTransaksi struct {
	Message             string       `json:"pesan_validate_transaksi"`
	DataReqeustMidtrans snap.Request `json:"data_request_midtrans_validate_transaksi"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Response Struct SnapTransaksi
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SnapTransaksi struct {
	SnapTransaksi *snap.Response        `json:"snap_response_token"`
	DataCheckout  []CheckoutData        `json:"checkout_data_hold"`
	DataAlamat    models.AlamatPengguna `json:"alamat_data_hold"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Response Struct Batal Transaksi
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ResponseBatalTransaksi struct {
	Message string `json:"pesan_batal_transaksi"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Response Struct Lock Transaksi
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ResponseLockTransaksi struct {
	Message string `json:"pesan_lock_transaksi"`
}

type ResponsePaidFailedTransaksi struct {
	Message string `json:"pesan_paid_failed_transaksi"`
}

// type ResponsePendingTransaksi struct {
// 	Message string `json:"pesan_pending_transaksi"`
// }
