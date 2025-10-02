package kurir_profiling_service

import "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"

type KredensialKurir struct {
	IDkurir       int64  `json:"id_kurir"`
	UsernameKurir string `json:"username_kurir"`
}

type PayloadPersonalProfilingKurir struct {
	DataKredensial KredensialKurir `json:"data_kredensial_kurir"`
	Username       string          `json:"ubah_username_kurir"`
	Nama           string          `json:"ubah_nama_kurir"`
	Email          string          `json:"ubah_email_kurir"`
}

type PayloadGeneralProfiling struct {
	DataIdentitas identity_kurir.IdentitasKurir `json:"data_identitas_kurir"`
	Deskripsi     string                        `json:"ubah_deskripsi_kurir"`
}
