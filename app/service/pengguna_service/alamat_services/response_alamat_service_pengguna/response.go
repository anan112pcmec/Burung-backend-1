package response_alamat_service_pengguna

type ResponseMembuatAlamat struct {
	Messages string `json:"pesan_membuat_alamat_pengguna"`
}

type ResponseHapusAlamat struct {
	Messages string `json:"pesan_hapus_alamat_pengguna"`
}
