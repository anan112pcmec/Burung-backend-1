package seller_order_processing_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"
)

type PayloadApproveOrder struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	DataTransaction []models.Transaksi             `json:"seller_transaksi_order_approve"`
}

type PayloadUnApproveOrder struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"data_identitas_seller"`
	Alasan          string                         `json:"alasan_order_unapprove"`
	DataTransaction []models.Transaksi             `json:"seller_transaksi_order_unapprove"`
}
