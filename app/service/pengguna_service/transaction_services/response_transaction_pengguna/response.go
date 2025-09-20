package response_transaction_pengguna

import "github.com/midtrans/midtrans-go/snap"

// ////////////////////////////////////////////////////////////////////////////////////
// CHECKOUT
// ////////////////////////////////////////////////////////////////////////////////////

type CheckoutData struct {
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

type ResponseDataCheckout struct {
	Message      string         `json:"pesan_chekout_barang"`
	DataResponse []CheckoutData `json:"data_response_checkout_barang"`
}

// ////////////////////////////////////////////////////////////////////////////////////
// TRANSAKSI
// ////////////////////////////////////////////////////////////////////////////////////

type ResponseDataValidateTransaksi struct {
	Message             string       `json:"pesan_validate_transaksi"`
	DataReqeustMidtrans snap.Request `json:"data_request_midtrans_validate_transaksi"`
}
