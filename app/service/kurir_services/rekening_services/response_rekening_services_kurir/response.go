package response_rekening_services_kurir

type ResponseMasukanRekeningKurir struct {
	Message string `json:"pesan_masukan_rekening_kurir"`
}

type ResponseEditRekeningKurir struct {
	Message string `json:"pesan_edit_rekening_kurir"`
}

type ResponseHapusRekeningKurir struct {
	Message string `json:"pesan_hapus_rekening_kurir"`
}
