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
		"status":                {"Online", "Offline"},
		"aksi_pengguna":         {"Registrasi", "Login", "Logout", "Pembelian", "Tambah Keranjang", "Hapus Keranjang", "Rating", "Update Profil", "Wishlist", "Pencarian"},
		"aksi_seler":            {"Registrasi", "Login", "Logout", "Pembelian", "Tambah Keranjang", "Hapus Keranjang", "Rating", "Update Profil", "Wishlist"},
		"jenis_seller":          {"Brands", "Distributors", "Personal"},
		"seller_dedication":     {"Pakaian & Fashion", "Kosmetik & Kecantikan", "Elektronik & Gadget", "Buku & Media", "Makanan & Minuman", "Ibu & Bayi", "Mainan", "Olahraga & Outdoor", "Otomotif & Sparepart", "Rumah Tangga", "Alat Tulis", "Perhiasan & Aksesoris", "Produk Digital", "Bangunan & Perkakas", "Musik & Instrumen", "Film & Broadcasting", "Semua Barang"},
		"jenis_layanan_kurir":   {"Reguler", "Express", "Ekonomi", "Sameday", "NextDay", "Cargo"},
		"status_keranjang":      {"Ready", "UnReady"},
		"nama_bank":             {"BCA", "BRI", "BNI", "MANDIRI", "BTN", "CIMB", "PERMATA"},
		"status_perizinan":      {"Pending", "Diizinkan", "Dilarang"},
		"jenis_kendaraan_kurir": {"Motor", "Mobil", "Truk", "Pickup", "Lainnya", "Unknown"},
		"roda_kendaraan_kurir":  {"2", "3", "4"},
		"status_kurir_narik":    {"Idle", "OnDelivery", "Off"},
		"status_jenis_seller":   {"Pending", "Confirmed", "Declined"},
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
		"status_transaksi": {"Dibayar", "Diproses", "Waiting", "Dikirim", "Selesai", "Dibatalkan"},
		// Dibayar adalah status default sebuah transaksi sampai seller melakukan approval.
		// Setelah transaksi di-approve oleh seller, status akan berubah menjadi "Diproses".
		// Status akan berubah lagi menjadi "Waiting" setelah seller memutuskan untuk mengirim barang.
		// Status menjadi "Dikirim" ketika seller sudah menyerahkan barang ke kurir.
		// Status menjadi "Selesai" ketika pengguna telah menerima barang dan mengonfirmasi bahwa transaksi telah selesai.
		// "Dibatalkan" digunakan ketika pengguna atau seller membatalkan transaksi, baik karena kesepakatan maupun sepihak.
		// Pembatalan hanya bisa dilakukan selama status masih "Dibayar".

		"status_pengiriman": {"Packaging", "Picked Up", "Diperjalanan", "Sampai", "Trouble"},
		// "Packaging" muncul di tabel pengiriman dan terjadi ketika status transaksi adalah "Diproses".
		// Status berubah menjadi "Picked Up" ketika status transaksi berubah menjadi "Dikirim".
		// Status menjadi "Diperjalanan" ketika kurir memperbarui status sendiri selama proses pengantaran.
		// Dalam status ini, kurir juga mengupdate catatan serta koordinat (latitude, longitude)
		// di tabel "jejak_pengiriman" yang merupakan child dari tabel "pengiriman".
		// Status "Sampai" akan dikonfirmasi oleh kurir saat barang tiba di tujuan.
		// Segala hal tak terduga seperti barang tidak sesuai, masalah di jalan, atau kerusakan barang
		// akan masuk ke status "Trouble". Namun untuk saat ini, kita berasumsi semua berjalan lancar.

		"status_paid_failed": {"Ditinjau", "Pending", "Batal", "Lanjut"},
		// "Ditinjau" berarti sistem sedang melakukan pemeriksaan terhadap transaksi gagal.
		// Kegagalan umumnya disebabkan oleh kesalahan foreign key, sehingga sistem akan melakukan self-healing data.
		// Setelah proses perbaikan (self-healing) selesai, status akan otomatis berubah menjadi "Pending".
		// Pada tahap "Pending", pengguna dapat memilih untuk melanjutkan atau membatalkan.
		// Secara default, jika tidak ada tindakan, status akan otomatis berubah menjadi "Batal" setelah 5 jam.
		// Jika pengguna memilih untuk melanjutkan, status berubah menjadi "Lanjut".
		// Dalam status "Lanjut", data akan dialihkan ke tabel transaksi dan pembayaran,
		// kemudian proses akan berlanjut seperti transaksi normal pada umumnya.
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
