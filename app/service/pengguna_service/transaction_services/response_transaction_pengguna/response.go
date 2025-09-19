package response_transaction_pengguna

type CheckoutData struct {
	IdBarangInduk    int32  `json:"id_barang_induk_keranjang"`
	IdKategoriBarang int64  `json:"id_kategori_barang_keranjang"`
	NamaBarang       string `json:"nama_barang_keranjang"`
	NamaKategori     string `json:"nama_kategori_barang_keranjang"`
	Dipesan          int32  `json:"dipesan_barang_keranjang"`
	Status           bool   `json:"status_barang_keranjang"`
	Message          string `json:"pesan_data_keranjang"`
}

type ResponseDataCheckout struct {
	Message      string         `json:"pesan_chekout_barang"`
	DataResponse []CheckoutData `json:"data_response_checkout_barang"`
}
