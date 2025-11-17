package pengguna_wishlist_services

import "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/identity_pengguna"

type PayloadTambahBarangKeWishlist struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	IdBarangInduk     int32                              `json:"id_barang_induk"`
}

type PayloadHapusBarangDariWishlist struct {
	IdentitasPengguna identity_pengguna.IdentityPengguna `json:"identitas_pengguna"`
	IdWishlist        int64                              `json:"id_wishlist"`
	IdBarangInduk     int32                              `json:"id_barang_induk"`
}
