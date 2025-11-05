package kurir_alamat_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"

type PayloadMasukanAlamatKurir struct {
	IdentitasKurir  identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	PanggilanAlamat string                        `json:"panggilan_alamat"`
	NomorTelephone  string                        `json:"nomor_telephone"`
	NamaAlamat      string                        `json:"nama_alamat"`
	Kota            string                        `json:"kota"`
	KodeNegara      string                        `json:"kode_negara"`
	KodePos         string                        `json:"kode_pos"`
	Deskripsi       string                        `json:"deskripsi"`
	Longtitude      float64                       `json:"longtitude"`
	Latitude        float64                       `json:"latitude"`
}

type PayloadEditAlamatKurir struct {
	IdentitasKurir  identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IDAlamatKurir   int64                         `json:"id_alamat_kurir"`
	PanggilanAlamat string                        `json:"panggilan_alamat"`
	NomorTelephone  string                        `json:"nomor_telephone"`
	NamaAlamat      string                        `json:"nama_alamat"`
	Kota            string                        `json:"kota"`
	KodeNegara      string                        `json:"kode_negara"`
	KodePos         string                        `json:"kode_pos"`
	Deskripsi       string                        `json:"deskripsi"`
	Longtitude      float64                       `json:"longtitude"`
	Latitude        float64                       `json:"latitude"`
}

type PayloadHapusAlamatKurir struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdAlamatKurir  int64                         `json:"id_alamat_kurir"`
}
