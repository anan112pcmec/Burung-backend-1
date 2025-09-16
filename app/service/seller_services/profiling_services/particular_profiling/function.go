package particular_profiling

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"

)

var wg sync.WaitGroup
var mu sync.Mutex

func UbahUsernameSeller(username string, db *gorm.DB, rds *redis.Client) *ResponseUbahUsername {
	var countUsername int64
	var saran []string

	if err := db.Model(&models.Seller{}).Where("username = ?", username).Count(&countUsername).Error; err != nil {
		return &ResponseUbahUsername{
			Message: "Server Bermasalah",
		}
	}

	retries := 0
	if countUsername > 0 {
		for len(saran) < 3 && retries < 20 {
			retries++
			wg.Add(1)
			go func() {
				defer wg.Done()

				usernameBaru := username + helper.GenerateRandomDigits()
				var tmp int64

				if err := db.Model(&models.Seller{}).Where("username = ?", usernameBaru).Count(&tmp).Error; err != nil {
					return
				}

				if tmp == 0 {
					mu.Lock()
					alreadyExists := false
					for _, v := range saran {
						if v == usernameBaru {
							alreadyExists = true
							break
						}
					}
					if !alreadyExists {
						saran = append(saran, usernameBaru)
					}
					mu.Unlock()
				}
			}()
		}

		wg.Wait()

		go func() {
			_ = db.Create(models.AktivitasSeller{Aksi: "Mengubah Username"})
		}()

		return &ResponseUbahUsername{
			Message: "Gagal, coba gunakan nama yang disarankan",
			Saran:   saran,
		}
	}

	if retries <= 20 {
		return &ResponseUbahUsername{
			Message: "Gagal, coba lagi nanti",
		}
	}

	return &ResponseUbahUsername{
		Message: "Berhasil",
	}
}

func UbahNamaSeller(nama string, db *gorm.DB) {
	
}

func UbahEmailSeller(ctx context.Context, email string, db *gorm.DB) {

}

func UbahJenisSeller(jenis string, db *gorm.DB) {

}

func UbahNorekSeller(ctx context.Context, norek string, db *gorm.DB) {

}

func UbahSellerDedicationSeller(seller_dedication string, db *gorm.DB) {

}

func UbahJamOperasionalSeller(jam_operasional string, db *gorm.DB) {

}

func UbahPunchlineSeller(punchline string, db *gorm.DB) {

}

func UbahPasswordSeller(ctx context.Context, password string, db *gorm.DB) {

}

func UbahDeskripsiSeller(deskripsi string, db *gorm.DB) {

}
