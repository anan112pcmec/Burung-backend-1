package migrate

import (
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
			}
		}
	}
}
