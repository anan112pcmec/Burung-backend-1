package kurir_rekening_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"

type PayloadMasukanRekeningKurir struct {
	IdentitasKurir  identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	NamaBank        string                        `json:"nama_bank"`
	NomorRekening   string                        `json:"nomor_rekening"`
	PemilikRekening string                        `json:"pemilik_rekening"`
}

type PayloadEditRekeningKurir struct {
	IdentitasKurir  identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdRekening      int64                         `json:"id_rekening"`
	NamaBank        string                        `json:"nama_bank"`
	NomorRekening   string                        `json:"nomor_rekening"`
	PemilikRekening string                        `json:"pemilik_rekening"`
}

type PayloadHapusRekeningKurir struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdRekening     int64                         `json:"id_rekening"`
}
