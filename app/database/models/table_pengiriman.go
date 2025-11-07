package models

import (
	"time"

	"gorm.io/gorm"
)

func (p *Pengiriman) BiayaKirimnya(untuk string) int16 {
	if untuk == "Sistem" {
		hasil := float64(p.BiayaKirim) / 0.2
		return int16(hasil)
	} else if untuk == "Kurir" {
		hasil := float64(p.BiayaKirim) / 0.8
		return int16(hasil)
	}

	return 0
}

type Pengiriman struct {
	ID                  int64          `gorm:"primaryKey;autoIncrement" json:"id_pengiriman"`
	IdTransaksi         int64          `gorm:"column:id_transaksi;not null" json:"id_transaksi_pengiriman"`
	Transaksi           Transaksi      `gorm:"foreignKey:IdTransaksi;references:ID" json:"-"`
	IdAlamatPengambilan int64          `gorm:"column:id_alamat_pengambilan;not null" json:"id_alamat_pengambilan_pengiriman"`
	AlamatGudang        AlamatGudang   `gorm:"foreignKey:IdAlamatPengambilan;references:ID" json:"-"`
	IdAlamatPengiriman  int64          `gorm:"column:id_alamat_pengiriman;not null" json:"id_alamat_pengiriman"`
	Alamat              AlamatPengguna `gorm:"foreignKey:IdAlamatPengiriman;references:ID" json:"-"`
	IdKurir             int64          `gorm:"column:id_kurir;not null" json:"id_kurir_pengiriman"`
	KendaraanRequired   string         `gorm:"column:kendaraan_required;type:jenis_kendaraan_kurir;not null" json:"kendaraan_required_pengiriman"`
	JenisPengiriman     string         `gorm:"column:jenis_pengiriman;not null;default:'reguler'" json:"jenis_pengiriman_transaksi"`
	Status              string         `gorm:"column:status;type:status_pengiriman;not null" json:"status_pengiriman"`
	BiayaKirim          int16          `gorm:"column:biaya_kirim;type:int2;not null;default:0" json:"biaya_kirim_pengiriman"`
	JarakTempuh         string         `gorm:"column:jarak_tempuh;type:varchar(100);not null" json:"jarak_tempuh_pengiriman"`
	KurirPaid           int32          `gorm:"column:kurir_paid;type:int4;not null;default:0" json:"kurir_paid_pengiriman"`
	BeratTotalKG        int16          `gorm:"column:berat_total_kg;type:int2;not null;default:0" json:"berat_total_kg_pengiriman"`
	CreatedAt           time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Pengiriman) TableName() string {
	return "pengiriman"
}

type JejakPengiriman struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id_jejak_pengiriman"`
	IdPengiriman int64          `gorm:"column:id_pengiriman;not null" json:"id_pengiriman_jejak_pengiriman"`
	Pengiriman   Pengiriman     `gorm:"foreignKey:IdPengiriman;references:ID"`
	Lokasi       string         `gorm:"column:lokasi;type:text;" json:"lokasi_jejak_pengiriman"`
	Keterangan   string         `gorm:"column:keterangan;type:text;not null;" json:"keterangan_jejak_pengiriman"`
	Latitude     float64        `gorm:"column:latitude;type:numeric(10,8);not null;" json:"latitude_jejak_pengiriman"`
	Longtitude   float64        `gorm:"column:longtitude;type:numeric(10,8);not null" json:"longtitude_jejak_pengiriman"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (JejakPengiriman) TableName() string {
	return "jejak_pengiriman"
}

type LayananPengirimanKurir struct {
	NamaLayanan  string `gorm:"column:nama_layanan;not null" json:"nama_layanan"`
	HargaLayanan int32  `gorm:"column:harga_layanan;type:int4;not null" json:"harga_layanan"`
}

func (LayananPengirimanKurir) TableName() string {
	return "layanan_pengiriman_kurir"
}

type Ongkir struct {
	Value int16  `gorm:"primaryKey"`
	Nama  string `gorm:"size:50;not null"`
}

func (Ongkir) TableName() string {
	return "ongkir"
}
