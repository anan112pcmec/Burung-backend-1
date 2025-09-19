package response_transaction_pengguna

type CheckoutData struct {
	NamaBarang   string `json:"nama_barang_keranjang"`
	NamaKategori string `json:"nama_kategori_barang_keranjang"`
	Dipesan      int32  `json:"dipesan_barang_keranjang"`
	Status       bool   `json:"status_barang_keranjang"`
	Message      string `json:"pesan_data_keranjang"`
}

type ResponseDataCheckout struct {
	Message      string         `json:"pesan_chekout_barang"`
	DataResponse []CheckoutData `json:"data_response_checkout_barang"`
}
