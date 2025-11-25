package seller_transaksi_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"

type PayloadApproveOrder struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdTransaksi     int64                          `json:"id_transaksi"`
	Catatan         string                         `json:"catatan_approve"`
}

type PayloadUnApproveOrder struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdTransaksi     int64                          `json:"id_transaksi"`
	Catatan         string                         `json:"catatan_unapprove"`
}
