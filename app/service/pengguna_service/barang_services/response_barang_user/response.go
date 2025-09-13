package response_barang_user

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

// ////////////////////////////////////////////////////////////////////////////
// MENGAMBIL DATA BARANG
// ////////////////////////////////////////////////////////////////////////////

type ResponseUserBarang struct {
	Harga string `json:"harga"`
	models.BarangInduk
}

type KategoriBarangDiambil struct {
	models.KategoriBarang
	Original bool `json:"barang_original"`
}

type ResponseUserBarangInduk struct {
	DataBarangInduk ResponseUserBarang      `json:"data_barang_induk"`
	DataKategori    []KategoriBarangDiambil `json:"data_kategori_barang_induk"`
}

// ////////////////////////////////////////////////////////////////////////////
// ENGAGEMENT USER BARANG
// ////////////////////////////////////////////////////////////////////////////

type ResponseLikesBarangUser struct {
	Message string `json:"pesan_likes_barang"`
}

type ResponseKomentarBarangUser struct {
	Message string `json:"pesan_komentar_barang"`
}

type ResponseEditKomentarBarangUser struct {
	Message string `json:"pesan_edit_komentar_barang"`
}
