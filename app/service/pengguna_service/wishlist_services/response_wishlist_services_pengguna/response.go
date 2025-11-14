package response_wishlist_services_pengguna

type ResponseTambahBarangKeWishlist struct {
	Message string `json:"pesan_tambah_barang_ke_wishlist"`
}

type ResponseHapusBarangDariWishlist struct {
	Message string `json:"pesan_hapus_barang_dari_wishlist"`
}
