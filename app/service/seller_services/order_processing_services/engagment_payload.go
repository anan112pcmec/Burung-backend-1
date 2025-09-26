package seller_order_processing_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

type PayloadApproveOrder struct {
	Seller          models.Seller      `json:"seller_credential_order_approve"`
	DataTransaction []models.Transaksi `json:"seller_transaksi_order_approve"`
}

type PayloadUnApproveOrder struct {
	Alasan          string             `json:"alasan_order_unapprove"`
	Seller          models.Seller      `json:"seller_credential_order_unapprove"`
	DataTransaction []models.Transaksi `json:"seller_transaksi_order_unapprove"`
}
