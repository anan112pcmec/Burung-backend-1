package response_barang_user

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

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
