package response_barang_service

// ////////////////////////////////////////////////////////////////////////////////
// BARANG INDUK
// ////////////////////////////////////////////////////////////////////////////////

type ResponseMasukanBarang struct {
	Message string `json:"pesan_memasukan_data_barang"`
}

type ResponseEditBarang struct {
	Message string `json:"pesan_edit_data_barang"`
}

type ResponseHapusBarang struct {
	Message string `json:"pesan_hapus_data_barang"`
}

// ////////////////////////////////////////////////////////////////////////////////
// KATEGORI BARANG
// ////////////////////////////////////////////////////////////////////////////////

type ResponseTambahKategori struct {
	Message string `json:"pesan_masukan_kategori_barang"`
}

type ResponseHapusKategori struct {
	Message string `json:"pesan_hapus_kategori_barang"`
}

type ResponseEditKategori struct {
	Message string `json:"pesan_edit_kategori_barang"`
}

// ////////////////////////////////////////////////////////////////////////////////
// STOK BARANG
// ////////////////////////////////////////////////////////////////////////////////

type ResponseEditStokBarang struct {
	Message string `json:"pesan_edit_stok_barang"`
}

type ResponseDownBarang struct {
	Message string `json:"pesan_down_stok_barang"`
}

type ResponseDownKategori struct {
	Message string `json:"pesan_down_stok_kategori_barang"`
}
