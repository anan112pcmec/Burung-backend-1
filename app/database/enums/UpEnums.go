package enums

import (
	"fmt"

	"gorm.io/gorm"
)

func UpEnumsEntity(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	enumMap := map[string][]string{
		"status":              {"Online", "Offline"},
		"aksi_pengguna":       {"Registrasi", "Login", "Logout", "Pembelian", "Tambah Keranjang", "Hapus Keranjang", "Rating", "Update Profil", "Wishlist", "Pencarian"},
		"aksi_seler":          {"Registrasi", "Login", "Logout", "Pembelian", "Tambah Keranjang", "Hapus Keranjang", "Rating", "Update Profil", "Wishlist"},
		"jenis_seller":        {"Brands", "Distributors", "Personal"},
		"seller_dedication":   {"Pakaian & Fashion", "Kosmetik & Kecantikan", "Elektronik & Gadget", "Buku & Media", "Makanan & Minuman", "Ibu & Bayi", "Mainan", "Olahraga & Outdoor", "Otomotif & Spaarepart", "Rumah Tangga", "Alat Tulis", "Perhiasan & Aksesoris", "Produk Digital", "Bangunan & Perkakas", "Musik & Instrumen", "Film & Broadcasting", "Semua Barang"},
		"jenis_layanan_kurir": {"Reguler", "Express", "Ekonomi", "Sameday", "NextDay", "Cargo"},
		"status_keranjang":    {"Ready", "UnReady"},
		"nama_bank":           {"BCA", "BRI", "BNI", "MANDIRI", "BTN", "CIMB", "PERMATA"},
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

func UpBarangEnums(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	enumMap := map[string][]string{
		"status_varian": {"Ready", "Dipesan", "Diproses", "Terjual", "Down"},
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
		"metode_pembayaran": {"Transfer Bank", "Kartu Kredit", "E Wallet", "COD"},
		"status_transaksi":  {"Dibayar", "Diproses", "Waiting", "Dikirim", "Selesai", "Dibatalkan"},
		"status_pengiriman": {"Packaging", "Picked Up", "Diperjalanan", "Sampai"},
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

	enumOngkir := []int16{13000, 17000, 10000, 31000, 25000, 7000}

	// Drop table dulu kalau ada
	dropSQL := `DROP TABLE IF EXISTS ongkir;`
	if err := tx.Exec(dropSQL).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create table baru
	createSQL := `CREATE TABLE ongkir (value SMALLINT PRIMARY KEY);`
	if err := tx.Exec(createSQL).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, val := range enumOngkir {
		insertSQL := fmt.Sprintf("INSERT INTO ongkir (value) VALUES (%d);", val)
		if err := tx.Exec(insertSQL).Error; err != nil {
			tx.Rollback()
			return err
		}
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
