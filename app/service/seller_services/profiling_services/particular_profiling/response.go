package seller_particular_profiling

type ResponseUbahUsername struct {
	Message string   `json:"pesan_ubah_username"`
	Saran   []string `json:"saran_username"`
}

type ResponseUbahNama struct {
	Message string `json:"pesan_ubah_nama"`
}

type ResponseUbahEmail struct {
	Message string `json:"pesan_ubah_email"`
}

type ResponseUbahJamOperasional struct {
	Message string `json:"pesan_ubah_jam_operasional"`
}

type ResponseUbahPunchline struct {
	Message string `json:"pesan_ubah_punchline"`
}

type ResponseUbahDeskripsi struct {
	Message string `json:"pesan_ubah_deskripsi"`
}
