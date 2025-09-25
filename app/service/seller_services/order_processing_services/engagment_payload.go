package seller_order_processing_services

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"

)

type PayloadApproveOrder struct {
	Seller          models.Seller      `json:"seller_credential_order_approve"`
	DataTransaction []models.Transaksi `json:"seller_transaksi_order_approve"`
}

type PayloadUnApproveOrder struct {
	Seller          models.Seller      `json:"seller_credential_order_unapprove"`
	DataTransaction []models.Transaksi `json:"seller_transaksi_order_unapprove"`
}
