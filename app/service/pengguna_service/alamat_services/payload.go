package pengguna_alamat_services

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type PayloadMasukanAlamatPengguna struct {
	DataAlamat models.AlamatPengguna `json:"data_alamat_pengguna"`
}

type PayloadHapusAlamatPengguna struct {
	IdPengguna      int64  `json:"id_pengguna_hapus_alamat"`
	PanggilanAlamat string `json:"panggilan_alamat_hapus_alamat"`
}
