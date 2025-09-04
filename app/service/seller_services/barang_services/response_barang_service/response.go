package response_barang_service

type ResponseMasukanBarang struct {
	Message string `json:"pesan_memasukan_data_barang"`
}

type ResponseEditBarang struct {
	Message string `json:"pesan_edit_data_varang"`
}
