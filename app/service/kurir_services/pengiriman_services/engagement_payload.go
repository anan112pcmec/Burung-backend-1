package kurir_pengiriman_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"
)

type PayloadAmbilPengiriman struct {
	Kredensial     identity_kurir.IdentitasKurir `json:"data_kredensial_kurir"`
	DataPengiriman []models.Pengiriman           `json:"data_pengiriman"`
}

type PayloadUpdatePengiriman struct {
	Kredensial          identity_kurir.IdentitasKurir `json:"data_informasi_kurir"`
	IDPengiriman        int64                         `json:"id_pengiriman"`
	DataJejakPengiriman models.JejakPengiriman        `json:"data_jejak_pengiriman"`
	StatusPengiriman    string                        `json:"status_pengiriman"`
}
