package response_etalase_services_seller

type ResponseMenambahEtalase struct {
	Message string `json:"pesan_menambah_etalase_seller"`
}

type ResponseEditEtalase struct {
	Message string `json:"pesan_edit_etalase_seller"`
}

type ResponseHapusEtalase struct {
	Message string `json:"pesan_hapus_etalase_seller"`
}

type ResponseTambahBarangKeEtalase struct {
	Message string `json:"pesan_tambah_barang_ke_etalase_seller"`
}

type ResponseHapusBarangKeEtalase struct {
	Message string `json:"pesan_hapus_barang_ke_etalase"`
}
