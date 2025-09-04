package seller_service

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type PayloadMasukanBarang struct {
	IdSeller       int32                   `json:"id_seller"`
	BarangInduk    models.BarangInduk      `json:"barang_induk_dimasukan"`
	KategoriBarang []models.KategoriBarang `json:"informasi_kategori"`
}

type PayloadEditBarang struct {
	IdSeller    int32              `json:"id_seller"`
	BarangInduk models.BarangInduk `json:"barang_induk_edit"`
}

type PayloadHapusBarang struct {
	IdSeller    int32              `json:"id_seller"`
	BarangInduk models.BarangInduk `json:"barang_induk_hapus"`
}
