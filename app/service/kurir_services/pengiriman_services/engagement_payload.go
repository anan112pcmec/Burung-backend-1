package kurir_pengiriman_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/kredensial_kurir"
)

type PayloadAmbilPengiriman struct {
	Kredensial     kredensial_kurir.Kredensial `json:"data_kredensial_kurir"`
	DataPengiriman []models.Pengiriman         `json:"data_pengiriman"`
}
