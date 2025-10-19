package response_barang

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type KategoriBarangDiambil struct {
	models.KategoriBarang
	Original bool `json:"barang_original"`
}

type ResponseUserBarangInduk struct {
	DataBarangInduk models.BarangInduk      `json:"data_barang_induk"`
	DataKategori    []KategoriBarangDiambil `json:"data_kategori_barang_induk"`
}
