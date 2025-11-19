package kurir_pengiriman_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/identity_kurir"

type PayloadAktifkanBid struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	Mode           string                        `json:"mode"`
	Alamat         string                        `json:"alamat"`
	Longitude      float64                       `json:"longitude"`
	Latitude       float64                       `json:"latitude"`
	MaxJarak       int16                         `json:"max_jarak"`
	MaxRadius      int16                         `json:"max_radius"`
	MaxKg          int16                         `json:"max_kg"`
}

type PayloadUpdateLokasiBidKurir struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdBidDataKurir int64                         `json:"id_bid_kurir"`
	Longitude      float64                       `json:"longitude"`
	Latitude       float64                       `json:"latitude"`
	Alamat         string                        `json:"alamat"`
}

type PayloadNonaktifkanBid struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
	IdBidDataKurir int64                         `json:"id_bid_kurir"`
}

type PayloadAmbilPengiriman struct {
	IdentitasKurir identity_kurir.IdentitasKurir `json:"identitas_kurir"`
}
