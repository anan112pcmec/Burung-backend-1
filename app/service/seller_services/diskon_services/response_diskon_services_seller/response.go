package response_diskon_services_seller

type ResponseTambahDiskonProduk struct {
	Message string `json:"pesan_tambah_diskon_produk"`
}

type ResponseEditDiskonProduk struct {
	Message string `json:"pesan_edit_diskon_produk"`
}

type ResponseHapusDiskonProduk struct {
	Message string `json:"pesan_hapus_diskon_produk"`
}

type ResponseTetapkanDiskonPadaBarang struct {
	Message string `json:"pesan_tetapkan_diskon_pada_barang"`
}

type ResponseHapusDiskonPadaBarang struct {
	Message string `json:"pesan_hapus_diskon_pada_barang"`
}
