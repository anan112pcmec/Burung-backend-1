package models

import (
	"time"

	"gorm.io/gorm"
)

type AlamatEkspedisi struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id_alamat_ekspedisi"`
	Kota       string         `gorm:"column:kota;type:nama_kota" json:"kota_alamat_ekspedisi"`
	NamaAlamat string         `gorm:"column:nama_alamat;type:text" json:"nama_alamat_ekspedisi"`
	Lokasi     string         `gorm:"column:lokasi;type:text;not null" json:"lokasi_alamat_ekspedisi"`
	Longitude  float64        `gorm:"column:longitude;type:numeric(11,8);not null" json:"longitude_alamat_ekspedisi"`
	Latitude   float64        `gorm:"column:latitude;type:numeric(11,8);not null" json:"latitude_alamat_ekspedisi"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (AlamatEkspedisi) TableName() string {
	return "alamat_ekspedisi"
}
