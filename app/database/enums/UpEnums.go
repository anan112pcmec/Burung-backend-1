package enums

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

func UpEnumsEntity(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	enumMap := map[string][]string{
		"status":                     {"Online", "Offline"},
		"aksi_pengguna":              {"Registrasi", "Login", "Logout", "Pembelian", "Tambah Keranjang", "Hapus Keranjang", "Rating", "Update Profil", "Wishlist", "Pencarian"},
		"aksi_seler":                 {"Registrasi", "Login", "Logout", "Pembelian", "Tambah Keranjang", "Hapus Keranjang", "Rating", "Update Profil", "Wishlist"},
		"jenis_seller":               {"Brands", "Distributors", "Personal"},
		"seller_dedication":          {"Pakaian & Fashion", "Kosmetik & Kecantikan", "Elektronik & Gadget", "Buku & Media", "Makanan & Minuman", "Ibu & Bayi", "Mainan", "Olahraga & Outdoor", "Otomotif & Sparepart", "Rumah Tangga", "Alat Tulis", "Perhiasan & Aksesoris", "Produk Digital", "Bangunan & Perkakas", "Musik & Instrumen", "Film & Broadcasting", "Semua Barang"},
		"jenis_layanan_kurir":        {"Reguler", "Express", "Ekonomi", "Sameday", "NextDay", "Cargo"},
		"status_keranjang":           {"Ready", "UnReady"},
		"nama_bank":                  {"BCA", "BRI", "BNI", "MANDIRI", "BTN", "CIMB", "PERMATA"},
		"status_perizinan_kendaraan": {"Pending", "Diizinkan", "Dilarang"},
		"jenis_kendaraan_kurir":      {"Motor", "Mobil", "Truk", "Pickup", "Lainnya", "Unknown"},
		"roda_kendaraan_kurir":       {"2", "3", "4"},
		"status_kurir_narik":         {"Idle", "OnDelivery", "Off"},
	}

	for enumName, values := range enumMap {
		var exists bool
		checkSQL := "SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = ?);"
		if err := tx.Raw(checkSQL, enumName).Scan(&exists).Error; err != nil {
			tx.Rollback()
			return err
		}

		if !exists {
			createSQL := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", enumName, joinWithQuotes(values))
			if err := tx.Exec(createSQL).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func UpBarangEnums(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	enumMap := map[string][]string{
		"status_varian": {"Ready", "Dipesan", "Diproses", "Terjual", "Down", "Pending"},
	}

	for enumName, values := range enumMap {
		// Cek apakah enum sudah ada
		var exists bool
		checkSQL := "SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = ?);"
		if err := tx.Raw(checkSQL, enumName).Scan(&exists).Error; err != nil {
			tx.Rollback()
			return err
		}

		if !exists {
			// Create type baru
			createSQL := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", enumName, joinWithQuotes(values))
			if err := tx.Exec(createSQL).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func UpEnumsTransaksi(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	enumMap := map[string][]string{
		"status_transaksi":  {"Dibayar", "Diproses", "Waiting", "Dikirim", "Selesai", "Dibatalkan"},
		"status_pengiriman": {"Packaging", "Picked Up", "Diperjalanan", "Sampai", "Trouble"},
	}

	for enumName, values := range enumMap {
		var exists bool
		checkSQL := "SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = ?);"
		if err := tx.Raw(checkSQL, enumName).Scan(&exists).Error; err != nil {
			tx.Rollback()
			return err
		}

		if !exists {
			createSQL := fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", enumName, joinWithQuotes(values))
			if err := tx.Exec(createSQL).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	enumOngkir := []models.Ongkir{
		{13000, "fast"},
		{17000, "express"},
		{10000, "reguler"},
		{31000, "sameday"},
		{25000, "instant"},
		{7000, "ekonomi"},
	}

	dropSQL := `DROP TABLE IF EXISTS ongkir;`
	if err := tx.Exec(dropSQL).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.AutoMigrate(&models.Ongkir{}); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Create(&enumOngkir).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func joinWithQuotes(values []string) string {
	res := ""
	for i, v := range values {
		res += fmt.Sprintf("'%s'", v)
		if i != len(values)-1 {
			res += ","
		}
	}
	return res
}
