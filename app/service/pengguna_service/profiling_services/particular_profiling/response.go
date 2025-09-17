package particular_profiling_pengguna

type ResponseUbahUsername struct {
	Message string   `json:"pesan_ubah_username_pengguna"`
	Saran   []string `json:"saran_username_pengguna"`
}

type ResponseUbahNama struct {
	Message string `json:"pesan_ubah_nama_pengguna"`
}

type ResponseUbahEmail struct {
	Message string `json:"pesan_ubah_email_pengguna"`
}
