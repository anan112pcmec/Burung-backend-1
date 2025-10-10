package response_seller

type ResponseRandomSeller struct {
	Message string `json:"pesan_ambil_random_seller"`
}

type ResponseAmbilNamaSeller struct {
	Message string `json:"pesan_ambil_nama_seller"`
}

type ResponseAmbilJenisSeller struct {
	Message string `json:"pesan_ambil_jenis_seller"`
}
