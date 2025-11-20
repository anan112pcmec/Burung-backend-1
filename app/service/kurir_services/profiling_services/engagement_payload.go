package kurir_profiling_service

import "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"

type PayloadPersonalProfilingKurir struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	Username       string                        `json:"ubah_username_kurir"`
	Nama           string                        `json:"ubah_nama_kurir"`
	Email          string                        `json:"ubah_email_kurir"`
}

type PayloadGeneralProfiling struct {
	DataIdentitas identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	Deskripsi     string                        `json:"ubah_deskripsi_kurir"`
}
