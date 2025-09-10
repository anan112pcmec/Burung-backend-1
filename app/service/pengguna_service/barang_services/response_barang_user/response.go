package response_barang_user

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type UserBarang struct {
	models.BarangInduk
	Harga string `json:"harga"`
}

type ResponseUserBarang struct {
	Barang []UserBarang
}
