package migrate

import (
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
)

func UpEntity(db *gorm.DB) {

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	wg.Add(3)

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Seller{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Seller ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Pengguna{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Pengguna ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Kurir{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Kurir ✅")
	}()

	wg.Wait()
	close(errCh)

	// cek apakah ada error
	for err := range errCh {
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	log.Println("All migrations completed successfully 🚀")
}

func UpBarang(db *gorm.DB) {
	if err := db.AutoMigrate(&models.BarangInduk{}); err == nil {
		log.Println("Migration Table Barang Induk Berhasil")
		if err1 := db.AutoMigrate(&models.KategoriBarang{}); err1 == nil {
			log.Println("Migration Table Kategori Barang")
			if err2 := db.AutoMigrate(&models.VarianBarang{}); err2 == nil {
				log.Println("Migration Table Varian Barang")
			} else {
				log.Fatalf("Gagal Membuat Table Varian Barang")
			}
		} else {
			log.Fatalf("Gagal Membuat Table KategoriBarang")
		}
	} else {
		fmt.Println("Gagal Migrasi Keseluruhan Barang")
	}
}

func UpTransaksi(db *gorm.DB) {
	if err := db.AutoMigrate(&models.Transaksi{}); err == nil {
		log.Println("Berhasil membuat Table Transaksi")
		if err1 := db.AutoMigrate(&models.Pembayaran{}); err1 == nil {
			log.Println("Berhasil Membuat Table Pembayaran")
		} else {
			log.Fatalf("Gagal Membuat Table Pembayaran")
		}
		if err2 := db.AutoMigrate(&models.Pengiriman{}); err2 == nil {
			log.Println("Berhasil Membuat Table Pengiriman")
			if err3 := db.AutoMigrate(&models.JejakPengiriman{}); err3 == nil {
				log.Println("Berhasil Membuat Table Jejak Pengiriman")
			}
		}
	} else {
		log.Fatalf("Gagal Membuat Keseluruhan Table transaksi")
	}
}

func UpEngagementEntity(db *gorm.DB) {
	var wg sync.WaitGroup
	errCh := make(chan error, 7)

	wg.Add(7)

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Komentar{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Komentar ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Keranjang{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Keranjang ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.BarangDisukai{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Barang Disukai ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Follower{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Follower ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.EntitySocialMedia{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Entity Social Media ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.AktivitasPengguna{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Aktivitas Pengguna ✅")
	}()

	go func() {
		defer wg.Done()
		if err := db.AutoMigrate(&models.Diskon{}); err != nil {
			errCh <- err
			return
		}
		log.Println("Migration success: Diskon ✅")
	}()

	wg.Wait()
	close(errCh)

	// cek apakah ada error
	for err := range errCh {
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}

	log.Println("All migrations completed successfully 🚀")
}
