package models

import (
	"time"

	"gorm.io/gorm"
)

type Pengiriman struct {
	ID                int64          `gorm:"primaryKey;autoIncrement" json:"id_pengiriman"`
	IdTransaksi       int64          `gorm:"column:id_transaksi;not null" json:"id_transaksi_pengiriman"`
	Transaksi         Transaksi      `gorm:"foreignKey:IdTransaksi;references:ID" json:"-"`
	IdSeller          int64          `gorm:"column:id_seller;not null" json:"id_seller_pengiriman"`
	Seller            Seller         `gorm:"foreignKey:IdSeller;references:ID" json:"-"`
	IdAlamatGudang    int64          `gorm:"column:id_alamat_gudang;not null" json:"id_alamat_gudang_pengiriman"`
	AlamatGudang      AlamatGudang   `gorm:"foreignKey:IdAlamatGudang;references:ID" json:"-"`
	IdAlamatPengguna  int64          `gorm:"column:id_alamat_pengguna;not null" json:"id_alamat_pengguna_pengiriman"`
	AlamatPengguna    AlamatPengguna `gorm:"foreignKey:IdAlamatPengguna;references:ID" json:"-"`
	IdKurir           *int64         `gorm:"column:id_kurir;not null;default:0" json:"id_kurir_pengiriman"`
	BeratBarang       int16          `gorm:"column:berat_barang;type:int2;not null" json:"berat_barang_pengiriman"`
	KendaraanRequired string         `gorm:"column:kendaraan_required;type:jenis_kendaraan_kurir;not null;default:'Motor'" json:"kendaraan_required_pengiriman"`
	JenisPengiriman   string         `gorm:"column:jenis_pengiriman;not null;type:jenis_layanan_kurir" json:"jenis_pengiriman"`
	JarakTempuh       string         `gorm:"column:jarak_tempuh;type:text;not null" json:"jarak_tempuh_pengiriman"`
	KurirPaid         int64          `gorm:"column:kurir_paid;type:int8;not null" json:"kurir_paid_pengiriman"`
	Status            string         `gorm:"column:status;type:status_pengiriman;not null;default:'Picked Up'" json:"status_pengiriman"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Pengiriman) TableName() string {
	return "pengiriman"
}

type JejakPengiriman struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" db:"id" json:"id_jejak_pengiriman"`
	IdPengiriman int64          `gorm:"column:id_pengiriman;not null" db:"id_pengiriman" json:"id_pengiriman_jejak_pengiriman"`
	Pengiriman   Pengiriman     `gorm:"foreignKey:IdPengiriman;references:ID" db:"-" json:"-"`
	Lokasi       string         `gorm:"column:lokasi;type:text" db:"lokasi" json:"lokasi_jejak_pengiriman"`
	Keterangan   string         `gorm:"column:keterangan;type:text;not null" db:"keterangan" json:"keterangan_jejak_pengiriman"`
	Latitude     float64        `gorm:"column:latitude;type:numeric(11,8);not null" db:"latitude" json:"latitude_jejak_pengiriman"`
	Longtitude   float64        `gorm:"column:longtitude;type:numeric(11,8);not null" db:"longtitude" json:"longtitude_jejak_pengiriman"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" db:"updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" db:"deleted_at" json:"deleted_at,omitempty"`
}

func (JejakPengiriman) TableName() string {
	return "jejak_pengiriman"
}

func (JejakPengiriman) HystoricalDbName() []string {
	return []string{"jejak_pengiriman"}
}

type PengirimanEkspedisi struct {
	ID                int64           `gorm:"primaryKey;autoIncrement" json:"id_pengiriman_ekspedisi"`
	IdTransaksi       int64           `gorm:"column:id_transaksi;not null" json:"id_transaksi_pengiriman_ekspedisi"`
	Transaksi         Transaksi       `gorm:"foreignKey:IdTransaksi;references:ID" json:"-"`
	IdSeller          int64           `gorm:"column:id_seller;not null" json:"id_seller_pengiriman_ekspedisi"`
	Seller            Seller          `gorm:"foreignKey:IdSeller;references:ID" json:"-"`
	IdAlamatGudang    int64           `gorm:"column:id_alamat_gudang;not null" json:"id_alamat_gudang_pengiriman_ekspedisi"`
	AlamatGudang      AlamatGudang    `gorm:"foreignKey:IdAlamatGudang;references:ID" json:"-"`
	IdAlamatEkspedisi int64           `gorm:"column:id_alamat_ekspedisi;not null" json:"id_alamat_ekspedisi_pengiriman_ekspedisi"`
	AlamatEkspedisi   AlamatEkspedisi `gorm:"foreignKey:IdAlamatEkspedisi;references:ID" json:"-"`
	IdKurir           *int64          `gorm:"column:id_kurir;not null;default:0" json:"id_kurir_pengiriman_ekspedisi"`
	BeratBarang       int16           `gorm:"column:berat_barang;type:int2;not null" json:"berat_barang_pengiriman_ekspedisi"`
	KendaraanRequired string          `gorm:"column:kendaraan_required;type:jenis_kendaraan_kurir;not null;default:'Motor'" json:"kendaraan_required_pengiriman_ekspedisi"`
	JenisPengiriman   string          `gorm:"column:jenis_pengiriman;type:jenis_layanan_kurir;not null" json:"jenis_pengiriman_ekspedisi"`
	JarakTempuh       string          `gorm:"column:jarak_tempuh;type:text;not null" json:"jarak_tempuh_pengiriman_ekspedisi"`
	KurirPaid         int64           `gorm:"column:kurir_paid;type:int8;not null" json:"kurir_paid_pengiriman_ekspedisi"`
	Status            string          `gorm:"column:status;type:status_pengiriman_ekspedisi;not null;default:'Picked Up'" json:"status_pengiriman_ekspedisi"`
}

func (PengirimanEkspedisi) TableName() string {
	return "pengiriman_ekspedisi"
}

type JejakPengirimanEkspedisi struct {
	ID                    int64               `gorm:"primaryKey;autoIncrement" json:"id_jejak_pengiriman"`
	IdPengirimanEkspedisi int64               `gorm:"column:id_pengiriman_ekspedisi;not null" json:"id_pengiriman_ekspedisi_jejak_pengiriman_ekspedisi"`
	Pengiriman            PengirimanEkspedisi `gorm:"foreignKey:IdPengirimanEkspedisi;references:ID"`
	Lokasi                string              `gorm:"column:lokasi;type:text" json:"lokasi_jejak_pengiriman"`
	Keterangan            string              `gorm:"column:keterangan;type:text;not null" json:"keterangan_jejak_pengiriman"`
	Latitude              float64             `gorm:"column:latitude;type:numeric(11,8);not null" json:"latitude_jejak_pengiriman"`
	Longitude             float64             `gorm:"column:longitude;type:numeric(11,8);not null" json:"longitude_jejak_pengiriman"`
	CreatedAt             time.Time           `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time           `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt             gorm.DeletedAt      `gorm:"index" json:"deleted_at,omitempty"`
}

func (JejakPengirimanEkspedisi) TableName() string {
	return "jejak_pengiriman_ekspedisi"
}
