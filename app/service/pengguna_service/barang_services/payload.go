package pengguna_service

import "github.com/anan112pcmec/Burung-backend-1/app/database/models"

type PayloadWatchBarang struct {
	ID int32 `json:"id_barang_induk"`
}

type PayloadLikesBarang struct {
	IDBarang int32 `json:"id_barang_induk_likes"`
	IDUser   int64 `json:"id_user_likes"`
}

type PayloadKomentarBarang struct {
	DataKomentar models.Komentar `json:"data_payload_komentar"`
}

type PayloadHapusKomentarBarang struct {
	IDKomentar int32 `json:"id_komentar_hapus_komentar"`
	IDEntity   int64 `json:"id_user_hapus_komentar"`
	Id_Barang  int32 `json:"id_barang_induk_hapus_komentar"`
}

type PayloadTambahDataKeranjangBarang struct {
	DataTambahKeranjang models.Keranjang `json:"data_payload_tambah_keranjang"`
}

type PayloadEditKomentarBarang struct {
	DataEditKomentar models.Komentar `json:"edit_payload_komentar"`
}
