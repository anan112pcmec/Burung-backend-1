package seller_etalase_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"

type PayloadMenambahEtalase struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	NamaEtalase     string                         `json:"nama"`
	Deskripsi       string                         `json:"deksripsi"`
}

type PayloadEditEtalase struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdEtalase       int64                          `json:"id_etalase"`
	NamaEtalase     string                         `json:"nama"`
	Deskripsi       string                         `json:"deksripsi"`
}

type PayloadHapusEtalase struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdEtalase       int64                          `json:"id_etalase"`
}

type PayloadTambahkanBarangKeEtalase struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdEtalase       int64                          `json:"id_etalase"`
	IdBarangInduk   int64                          `json:"id_barang_induk"`
}

type PayloadHapusBarangDiEtalase struct {
	IdentitasSeller   identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangKeEtalase int64                          `json:"id_barang_ke_etalase"`
	IdEtalase         int64                          `json:"id_etalase"`
	IdBarangInduk     int64                          `json:"id_barang_induk"`
}
