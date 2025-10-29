package seller_service

import (
	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/identity_seller"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Masukan Barang Induk
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadMasukanBarangInduk struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	BarangInduk     models.BarangInduk             `json:"data_barang_induk"`
	KategoriBarang  []models.KategoriBarang        `json:"data_kategori_barang_induk"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Edit Barang Induk
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadEditBarangInduk struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	BarangInduk     models.BarangInduk             `json:"barang_induk_edit"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Hapus Barang Induk
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadHapusBarangInduk struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	BarangInduk     models.BarangInduk             `json:"barang_induk_hapus"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Tambah Kategori
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadTambahKategori struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk   int32                          `json:"id_barang_induk_tambah_kategori"`
	KategoriBarang  []models.KategoriBarang        `json:"tambah_kategori_barang"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Edit Kategori
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadEditKategori struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk   int32                          `json:"id_barang_induk_edit_kategori"`
	KategoriBarang  []models.KategoriBarang        `json:"edit_kategori_barang"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Hapus Kategori
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadHapusKategori struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk   int32                          `json:"id_barang_induk_hapus_kategori"`
	KategoriBarang  []models.KategoriBarang        `json:"hapus_kategori_barang"`
}

type IdKategoriDanStok struct {
	IdKategoriBarang   int64  `json:"id_kategori_barang_edit_stok"`
	NamaKategoriBarang string `json:"nama_kategori_barang_edit_stok"`
	SkuKategoriBarang  string `json:"sku_kategori_barang"`
	JumlahStok         int32  `json:"jumlah_stok"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Edit Stok Barang
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadEditStokBarang struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk   int32                          `json:"id_barang_induk_stok_edit"`
	Barang          []IdKategoriDanStok            `json:"stok_barang_edit"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Down Barang Induk
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadDownBarangInduk struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk   int32                          `json:"id_barang_induk_down"`
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Payload Struct Down Kategori Barang
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PayloadDownKategoriBarang struct {
	IdentitasSeller  identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk    int32                          `json:"id_barang_induk_down"`
	IdKategoriBarang int64                          `json:"id_kategori_barang_down"`
}

// ////////////////////////////////////////////////////////////////////////////////
// Payload Struct Edit Rekening Barang Induk
// ////////////////////////////////////////////////////////////////////////////////

type PayloadEditRekeningBarangInduk struct {
	IdentitasSeller  identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk    int32                          `json:"id_barang_induk"`
	IdRekeningSeller int64                          `json:"id_rekening_seller"`
}

// ////////////////////////////////////////////////////////////////////////////////
// Payload Struct Edit Alamat Barang Induk
// ////////////////////////////////////////////////////////////////////////////////

type PayloadEditAlamatBarangInduk struct {
	IdentitasSeller identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk   int32                          `json:"id_barang_induk"`
	IdAlamatGudang  int64                          `json:"id_alamat_gudang"`
}

// ////////////////////////////////////////////////////////////////////////////////
// Payload Struct Edit Alamat Barang Kategori
// ////////////////////////////////////////////////////////////////////////////////

type PayloadEditAlamatBarangKategori struct {
	IdentitasSeller  identity_seller.IdentitySeller `json:"identitas_seller"`
	IdBarangInduk    int32                          `json:"id_barang_induk"`
	IdKategoriBarang int64                          `json:"id_kategori_barang"`
	IdAlamatGudang   int64                          `json:"id_alamat_gudang"`
}
