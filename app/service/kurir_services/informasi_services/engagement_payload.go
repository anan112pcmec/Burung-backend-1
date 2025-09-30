package kurir_informasi_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"
)

type PayloadInformasiDataKendaraan struct {
	DataIdentitasKurir     identity_kurir.IdentitasKurir  `json:"identitas_kurir"`
	DataInformasiKendaraan models.InformasiKendaraanKurir `json:"informasi_kendaraan"`
}

type PayloadEditInformasiDataKendaraan struct {
	DataIdentitasKurir     identity_kurir.IdentitasKurir  `json:"identitas_kurir"`
	DataInformasiKendaraan models.InformasiKendaraanKurir `json:"edit_informasi_kendaraan"`
}

type PayloadInformasiDataKurir struct {
	DataIdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	DataInformasiKurir models.InformasiKurir         `json:"informasi_kurir"`
}

type PayloadEditInformasiDataKurir struct {
	DataIdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	DataInformasiKurir models.InformasiKurir         `json:"edit_informasi_kurir"`
}
