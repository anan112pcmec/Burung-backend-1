package models

import (
	"time"

	"gorm.io/gorm"
)

type Pembayaran struct {
	ID              int64          `gorm:"primaryKey;autoIncrement" json:"id_pembayaran"`
	IdPengguna      int64          `gorm:"column:id_pengguna;not null" json:"id_pengguna"`
	Pengguna        Pengguna       `gorm:"foreignKey:IdPengguna;references:ID" json:"-"`
	KodeTransaksiPG string         `gorm:"column:kode_transaksi_pg;not null" json:"kode_transaksi_pg_pembayaran"`
	KodeOrderSistem string         `gorm:"column:kode_order_sistem;type:varchar(250);unique;not null" json:"kode_order_sistem_pembayaran"`
	Provider        string         `gorm:"column:provider;type:text;not null;default:''" json:"provider_pembayaran"`
	Total           int32          `gorm:"column:total;type:int4;not null;default:0" json:"total_pembayaran"`
	PaymentType     string         `gorm:"column:payment_type;type:varchar(120);not null" json:"payment_type_pembayaran"`
	PaidAt          string         `gorm:"column:paid_at;type:text;not null;default:''" json:"paid_at_pembayaran"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Pembayaran) TableName() string {
	return "pembayaran"
}

type Transaksi struct {
	ID                  int64          `gorm:"primaryKey;autoIncrement" json:"id_transaksi"`
	IdPengguna          int64          `gorm:"column:id_pengguna;not null" json:"id_pengguna_transaksi"`
	Pengguna            Pengguna       `gorm:"foreignKey:IdPengguna;references:ID" json:"-"`
	IdSeller            int32          `gorm:"column:id_seller;not null" json:"id_seller_transaksi"`
	Seller              Seller         `gorm:"foreignKey:IdSeller;references:ID" json:"-"`
	IdBarangInduk       int64          `gorm:"column:id_barang_induk;not null" json:"id_barang_induk_transaksi"`
	BarangInduk         BarangInduk    `gorm:"foreignKey:IdBarangInduk;references:ID" json:"-"`
	IdKategoriBarang    int64          `gorm:"column:id_kategori_barang;not null" json:"id_kategori_barang_transaksi"`
	KategoriBarang      KategoriBarang `gorm:"foreignKey:IdKategoriBarang;references:ID" json:"-"`
	IdAlamatPengguna    int64          `gorm:"column:id_alamat_pengguna;not null" json:"id_alamat_pengguna_transaksi"`
	AlamatPengguna      AlamatPengguna `gorm:"foreignKey:IdAlamatPengguna;references:ID" json:"-"`
	IdAlamatGudang      int64          `gorm:"column:id_alamat_gudang;type:int8;not null" json:"id_alamat_gudang_transaksi"`
	IdAlamatEkspedisi   int64          `gorm:"column:id_alamat_ekspedisi;type:int8;not null" json:"id_alamat_ekspedisi"`
	IdPembayaran        int64          `gorm:"column:id_pembayaran;not null" json:"id_pembayaran_transaksi"`
	Pembayaran          Pembayaran     `gorm:"foreignKey:IdPembayaran;references:ID" json:"-"`
	KendaraanPengiriman string         `gorm:"column:kendaraan_pengiriman;type:jenis_kendaraan_kurir;default:'Motor';not null" json:"kendaraan_pengiriman_transaksi"`
	JenisPengiriman     string         `gorm:"column:jenis_pengiriman;type:jenis_layanan_kurir;not null" json:"jenis_pengiriman_transaksi"`
	JarakTempuh         string         `gorm:"column:jarak_tempuh;not null" json:"jarak_tempuh_transaksi"`
	BeratTotalKg        int16          `gorm:"column:berat_total_kg;type:int2;not null" json:"berat_total_kg_pengiriman_transaksi"`
	KodeOrderSistem     string         `gorm:"column:kode_order_sistem;type:varchar(100);not null" json:"kode_order_sistem_transaksi"`
	KodeResiEkspedisi   *string        `gorm:"column:kode_resi_ekspedisi;type:varchar(100)" json:"kode_resi_ekspedisi_transaksi"`
	Status              string         `gorm:"column:status;type:status_transaksi;default:'Dibayar';not null" json:"status_transaksi"`
	DibatalkanOleh      *string        `gorm:"column:dibatalkan_oleh;type:jenis_entity" json:"dibatalkan_oleh_transaksi"`
	Catatan             string         `gorm:"column:catatan;type:text" json:"catatan_transaksi"`
	KuantitasBarang     int32          `gorm:"column:kuantitas_barang;type:int4;not null" json:"kuantitas_barang"`
	IsEkspedisi         bool           `gorm:"column:is_ekspedisi;not null;default:false" json:"is_ekspedisi_transaksi"`
	SellerPaid          int64          `gorm:"column:seller_paid;type:int8;not null" json:"seller_paid_transaksi"`
	KurirPaid           int64          `gorm:"column:kurir_paid;type:int8;not null" json:"kurir_paid_transaksi"`
	EkspedisiPaid       int64          `gorm:"column:ekspedisi_paid;type:int8;not null" json:"ekspedisi_paid_transaksi"`
	Total               int64          `gorm:"column:total;type:int8;not null" json:"total_transaksi"`
	Reviewed            bool           `gorm:"column:reviewed;type:bool;not null;default:false" json:"reviewed_transaksi"`
	CreatedAt           time.Time      `gorm:"autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

func (Transaksi) TableName() string {
	return "transaksi"
}

type PembayaranFailed struct {
	ID                int64          `gorm:"primaryKey;autoIncrement" json:"id_paid_failed"`
	IdPengguna        int64          `gorm:"column:id_pengguna;not null" json:"id_pengguna"`
	Pengguna          Pengguna       `gorm:"foreignKey:IdPengguna;references:ID" json:"-"`
	FinishRedirectUrl string         `gorm:"column:finish_redirect_url" json:"finish_redirect_url"`
	FraudStatus       string         `gorm:"column:fraud_status" json:"fraud_status"`
	GrossAmount       string         `gorm:"column:gross_amount" json:"gross_amount"`
	OrderId           string         `gorm:"column:order_id" json:"order_id"`
	PaymentType       string         `gorm:"column:payment_type" json:"payment_type"`
	PdfUrl            string         `gorm:"column:pdf_url" json:"pdf_url"`
	StatusCode        string         `gorm:"column:status_code" json:"status_code"`
	StatusMessage     string         `gorm:"column:status_message" json:"status_message"`
	TransactionId     string         `gorm:"column:transaction_id" json:"transaction_id"`
	TransactionStatus string         `gorm:"column:transaction_status" json:"transaction_status"`
	TransactionTime   string         `gorm:"column:transaction_time" json:"transaction_time"`
	Bank              string         `gorm:"column:bank" json:"bank"`
	VaNumber          string         `gorm:"column:va_number" json:"va_number"`
	PaymentCode       string         `gorm:"column:payment_code" json:"payment_code"`
	Status            string         `gorm:"column:status;type:status_paid_failed;default:'Pending'" json:"status"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (PembayaranFailed) TableName() string {
	return "pembayaran_failed"
}

type TransaksiFailed struct {
	ID                 int64            `gorm:"primaryKey;autoIncrement" json:"id_transaksi_failed"`
	IdPembayaranFailed int64            `gorm:"column:id_pembayaran_failed;not null" json:"id_pembayaran_failed"`
	PembayaranFailed   PembayaranFailed `gorm:"foreignKey:IdPembayaranFailed;references:ID" json:"-"`
	IdPengguna         int64            `gorm:"column:id_pengguna;not null" json:"id_pengguna"`
	Pengguna           Pengguna         `gorm:"foreignKey:IdPengguna;references:ID" json:"-"`
	IdSeller           int32            `gorm:"column:id_seller;not null" json:"id_seller"`
	Seller             Seller           `gorm:"foreignKey:IdSeller;references:ID" json:"-"`
	IdBarangInduk      int32            `gorm:"column:id_barang_induk;not null" json:"id_barang_induk"`
	BarangInduk        BarangInduk      `gorm:"foreignKey:IdBarangInduk;references:ID" json:"-"`
	IdKategoriBarang   int64            `gorm:"id_kategori_barang;not null" json:"id_kategori_barang"`
	KategoriBarang     KategoriBarang   `gorm:"foreignKey:IdKategoriBarang;references:ID" json:"-"`
	IdAlamat           int64            `gorm:"column:id_alamat_pengguna" json:"id_alamat_pengguna"`
	JenisPengiriman    string           `gorm:"column:jenis_pengiriman;not null;default:'reguler'" json:"jenis_pengiriman"`
	Catatan            string           `gorm:"column:catatan;type:text" json:"catatan_transaksi"`
	Kuantitas          int16            `gorm:"column:kuantitas_barang;type:int2;not null" json:"kuantitas_barang"`
	Total              int64            `gorm:"column:total;type:int8;not null" json:"total"`
	CreatedAt          time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty"`
}

func (TransaksiFailed) TableName() string {
	return "transaksi_failed"
}
