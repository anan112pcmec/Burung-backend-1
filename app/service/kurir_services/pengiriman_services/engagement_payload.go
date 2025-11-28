package kurir_pengiriman_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"
)

type PayloadAktifkanBidKurir struct {
	IdentitasKurir  identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	JenisPengiriman string                        `json:"jenis_pengiriman"`
	Mode            string                        `json:"mode"`
	Provinsi        string                        `json:"provinsi"`
	Kota            string                        `json:"kota"`
	Alamat          string                        `json:"alamat"`
	Longitude       float64                       `json:"longitude"`
	Latitude        float64                       `json:"latitude"`
	MaxKg           int32                         `json:"max_kg"`
}

type PayloadUpdatePosisiBid struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdBidKurir     int64                         `json:"id_bid_kurir"`
	Longitude      float64                       `json:"longitude"`
	Latitude       float64                       `json:"latitude"`
}

type PayloadNonaktifkanBidKurir struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdBidKurir     int64                         `json:"id_bid_kurir"`
}
