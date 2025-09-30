package response_informasi_services_kurir

type ResponseAjukanInformasiKendaraan struct {
	Message string `json:"pesan_ajukan_informasi_kendaraan"`
}

type ResponseEditInformasiKendaraan struct {
	Message string `json:"pesan_edit_informasi_kendaraan"`
}

type ResponseAjukanInformasiKurir struct {
	Message string `json:"pesan_ajukan_informasi_kurir"`
}

type ResponseEditInformasiKurir struct {
	Message string `json:"pesan_edit_informasi_kurir"`
}
