package response_pengiriman_services_kurir

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type AmbilPengirimanDetail struct {
	DataPengiriman models.Pengiriman `json:"data_pengiriman"`
	Status         bool              `json:"status_ambil_pengriman"`
}

type ResponseAmbilPengiriman struct {
	Message       string                  `json:"pesan_ambil_pengiriman_kurir"`
	HasilResponse []AmbilPengirimanDetail `json:"response_ambil_pengiriman"`
}
