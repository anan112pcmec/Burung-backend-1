package seller_alamat_services

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type PayloadMasukanAlamatSeller struct {
	Data models.AlamatSeller `json:"data_alamat_seller"`
}

type PayloadHapusAlamatSeller struct {
	IDSeller        int32  `json:"id_seller_hapus_alamat"`
	PanggilanAlamat string `json:"panggilan_alamat_hapus_alamat"`
}
