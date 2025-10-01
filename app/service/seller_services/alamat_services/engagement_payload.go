package seller_alamat_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"
)

type PayloadTambahAlamatGudang struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	Data            models.AlamatGudang            `json:"data_tambah_alamat_gudang"`
}

type PayloadEditAlamatGudang struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	Data            models.AlamatGudang            `json:"data_edit_alamat_gudang"`
}

type PayloadHapusAlamatGudang struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	IdGudang        int64                          `json:"id_alamat_gudang"`
}
