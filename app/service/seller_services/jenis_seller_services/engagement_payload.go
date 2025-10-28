package jenis_seller_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Ajukan Ubah Jenis Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadAjukanUbahJenisSeller struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	DataDiajukan    models.Jenis_Seller            `json:"data_diajukan_ubah_jenis_seller"`
}
