package response_alamat_services_seller

type ResponseTambahAlamatGudang struct {
	Message string `json:"pesan_tambah_alamat_gudang"`
}

type ResponseEditAlamatGudang struct {
	Message string `json:"pesan_edit_alamat_gudang"`
}

type ResponseHapusAlamatGudang struct {
	Message string `json:"pesan_hapus_alamat_gudang"`
}
