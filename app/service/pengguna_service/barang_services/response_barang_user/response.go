package response_barang_user

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type ResponseUserBarang struct {
	Harga string `json:"harga"`
	models.BarangInduk
}
