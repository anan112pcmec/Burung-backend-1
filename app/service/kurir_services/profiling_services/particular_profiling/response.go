package particular_profiling_kurir

type ResponseUbahNama struct {
	Message string `json:"pesan_ubah_nama"`
}

type ResponseUbahUsername struct {
	Message       string   `json:"pesan_ubah_username"`
	SaranUsername []string `json:"saran_username"`
}

type ResponseUbahDeskripsi struct {
	Message string `json:"pesan_ubah_deskripsi"`
}

type ResponseUbahGmail struct {
	Message string `json:"pesan_ubah_gmail"`
}
