package response_barang

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
	Message string `json:"pesan_tambah_komentar_barang"`
}

type ResponseEditKomentarBarangUser struct {
	Message string `json:"pesan_edit_komentar_barang"`
}

type ResponseHapusKomentarBarangUser struct {
	Message string `json:"pesan_hapus_komentar_barang"`
}

type ResponseTambahKeranjangUser struct {
	Message string `json:"pesan_tambah_keranjang_barang"`
}

type ResponseEditKeranjangUser struct {
	Message string `json:"pesan_edit_keranjang_barang"`
}

type ResponseHapusKeranjangUser struct {
	Message string `json:"pesan_hapus_keranjang_barang"`
}
