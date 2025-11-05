package response_alamat_service_kurir

type ResponseMasukanAlamatKurir struct {
	Message string `json:"pesan_masukan_alamat_kurir"`
}

type ResponseEditAlamatKurir struct {
	Message string `json:"pesan_edit_alamat_kurir"`
}

type ResponseHapusAlamatKurir struct {
	Message string `json:"pesan_hapus_alamat_kurir"`
}
