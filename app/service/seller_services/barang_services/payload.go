package seller_service

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

// ////////////////////////////////////////////////////////////////////////////////
// BARANG INDUK
// ////////////////////////////////////////////////////////////////////////////////

type PayloadMasukanBarang struct {
	IdSeller       int32                   `json:"id_seller"`
	BarangInduk    models.BarangInduk      `json:"barang_induk_dimasukan"`
	KategoriBarang []models.KategoriBarang `json:"kategori_barang_induk"`
}

type PayloadEditBarang struct {
	IdSeller    int32              `json:"id_seller"`
	BarangInduk models.BarangInduk `json:"barang_induk_edit"`
}

type PayloadHapusBarang struct {
	IdSeller    int32              `json:"id_seller"`
	BarangInduk models.BarangInduk `json:"barang_induk_hapus"`
	JWT         string             `json:"jwt_key"`
}

// ////////////////////////////////////////////////////////////////////////////////
// KATEGORI BARANG
// ////////////////////////////////////////////////////////////////////////////////

type PayloadTambahKategori struct {
	IdBarangInduk  int32                   `json:"id_barang_induk_tambah_kategori"`
	IdSeller       int32                   `json:"id_seller_tambah_kategori"`
	KategoriBarang []models.KategoriBarang `json:"tambah_kategori_barang"`
}

type PayloadHapusKategori struct {
	IdBarangInduk  int32                   `json:"id_barang_induk_hapus_kategori"`
	IdSeller       int32                   `json:"id_seller_hapus_kategori"`
	KategoriBarang []models.KategoriBarang `json:"hapus_kategori_barang"`
}

type PayloadEditKategori struct {
	IdBarangInduk  int32                   `json:"id_barang_induk_edit_kategori"`
	IdSeller       int32                   `json:"id_seller_edit_kategori"`
	KategoriBarang []models.KategoriBarang `json:"edit_kategori_barang"`
}

// ////////////////////////////////////////////////////////////////////////////////
// STOK BARANG
// ////////////////////////////////////////////////////////////////////////////////

type IdKategoriDanStok struct {
	IdKategoriBarang   int64  `json:"id_kategori_barang_edit_stok"`
	NamaKategoriBarang string `json:"nama_kategori_barang_edit_stok"`
	SkuKategoriBarang  string `json:"sku_kategori_barang"`
	JumlahStok         int32  `json:"jumlah_stok"`
}

type PayloadEditStokBarang struct {
	IdBarangInduk int32               `json:"id_barang_induk_stok_edit"`
	IdSeller      int32               `json:"id_seller_edit_stok"`
	Barang        []IdKategoriDanStok `json:"stok_barang_edit"`
}

type PayloadDownBarangInduk struct {
	IdBarangInduk int32 `json:"id_barang_induk_down"`
	IdSeller      int32 `json:"id_seller_barang_induk_down"`
}

type PayloadDownKategoriBarang struct {
	IdBarangInduk    int32 `json:"id_barang_induk_down"`
	IdKategoriBarang int64 `json:"id_kategori_barang_down"`
	IdSeller         int32 `json:"id_seller_barang_induk_down"`
}
