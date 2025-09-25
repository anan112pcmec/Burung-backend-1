package response_order_processing_seller

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type ApprovedStatus struct {
	DataApproved models.Transaksi `json:"data_transaksi_approved"`
	Status       bool             `json:"status_approved"`
}

type UnApprovedStatus struct {
	DataUnApproved models.Transaksi `json:"data_transaksi_unapproved"`
	Status         bool             `json:"status_unapproved"`
}

type ResponseApproveTransaksiSeller struct {
	Message string            `json:"pesan_approve_transaksi_seller"`
	Hasil   *[]ApprovedStatus `json:"rincian_approved_transaksi"`
}

type ResponseUnApproveTransaksiSeller struct {
	Message string              `json:"pesan_unapprove_transaksi_seller"`
	Hasil   *[]UnApprovedStatus `json:"rincian_unapproved_transaksi"`
}
