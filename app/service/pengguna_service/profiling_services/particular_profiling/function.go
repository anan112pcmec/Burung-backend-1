package particular_profiling_pengguna

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
)

func UbahUsernamePengguna(db *gorm.DB, id_pengguna int64, username string) *ResponseUbahUsername {

	var countUsername int64
	saran := make([]string, 0, 3)

	if err := db.Model(&models.Pengguna{}).
		Where("username = ?", username).
		Count(&countUsername).Error; err != nil {
		log.Printf("[ERROR] Gagal memeriksa username: %v", err)
		return &ResponseUbahUsername{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
		}
	}

	if countUsername == 0 {
		if err_seller := db.Model(&models.Pengguna{}).
			Where("id = ?", id_pengguna).
			Update("username", username).Error; err_seller == nil {
			log.Printf("[INFO] Username berhasil diubah untuk pengguna ID %d", id_pengguna)
			return &ResponseUbahUsername{
				Message: "Username berhasil diubah.",
			}
		}
	}

	if countUsername > 0 {

		for len(saran) < 4 {

			usernameBaru := username + helper.GenerateRandomDigits()
			var tmp int64

			if err := db.Model(&models.Pengguna{}).
				Where("username = ?", usernameBaru).
				Count(&tmp).Error; err != nil {
				log.Printf("[WARN] Gagal generate saran username: %v", err)
				continue
			}

			if tmp == 0 {
				for _, v := range saran {
					if v == usernameBaru {
						continue
					}
				}
				saran = append(saran, usernameBaru)
			}

		}

		go func() {
			_ = db.Create(&models.AktivitasPengguna{
				Aksi: "Mengubah Username",
			})
		}()

		log.Printf("[WARN] Username '%s' sudah digunakan. Menyediakan saran.", username)
		return &ResponseUbahUsername{
			Message: "Username sudah digunakan. Silakan pilih salah satu saran berikut.",
			Saran:   saran,
		}
	}

	log.Printf("[ERROR] Gagal mengubah username untuk pengguna ID %d", id_pengguna)
	return &ResponseUbahUsername{
		Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
	}
}

func UbahNamaPengguna(id_pengguna int64, nama string, db *gorm.DB) *ResponseUbahNama {

	if err_db := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: id_pengguna}).Update("nama", nama).Error; err_db == nil {
		log.Printf("[INFO] Nama berhasil diubah untuk pengguna ID %d", id_pengguna)
		return &ResponseUbahNama{
			Message: "Nama berhasil diubah.",
		}
	}

	log.Printf("[ERROR] Gagal mengubah nama untuk pengguna ID %d", id_pengguna)
	return &ResponseUbahNama{
		Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
	}
}

func UbahEmailSeller(ctx context.Context, id_pengguna int64, email string, db *gorm.DB) *ResponseUbahEmail {
	sudah_ada := ""
	if err_sama := db.Model(models.Pengguna{}).Select("email").Where(models.Pengguna{ID: id_pengguna, Email: email}).First(&sudah_ada).Error; err_sama != nil {
		log.Printf("[INFO] Email baru belum digunakan oleh pengguna ID %d", id_pengguna)
	}

	if sudah_ada != "" {
		log.Printf("[WARN] Email yang dimasukkan sama dengan email lama untuk pengguna ID %d", id_pengguna)
		return &ResponseUbahEmail{
			Message: "Email yang dimasukkan sama dengan email sebelumnya.",
		}
	}

	var email_lama string
	if err_email := db.Model(models.Pengguna{}).Select("email").Where(models.Pengguna{ID: id_pengguna}).First(&email_lama).Error; err_email != nil {
		log.Printf("[ERROR] Gagal mendapatkan email lama pengguna ID %d: %v", id_pengguna, err_email)
		return &ResponseUbahEmail{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
		}
	}

	if err_ubah_email := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: id_pengguna}).Update("email", email).Error; err_ubah_email != nil {
		log.Printf("[ERROR] Gagal mengubah email untuk pengguna ID %d: %v", id_pengguna, err_ubah_email)
		return &ResponseUbahEmail{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
		}
	}

	go func() {
		to := []string{email_lama}
		cc := []string{}
		subject := "Pemberitahuan Pembaruan Email"
		message := fmt.Sprintf("Akun Burung Anda telah diubah email-nya pada %s menjadi %s. Mulai sekarang semua pemberitahuan akan dikirim ke email baru tersebut.", time.Now().Format("02-01-2006 15:04:05"), email)
		err := emailservices.SendMail(to, cc, subject, message)
		if err != nil {
			log.Printf("[ERROR] Gagal mengirim email notifikasi perubahan email ke %s: %v", email_lama, err)
		} else {
			log.Printf("[INFO] Notifikasi perubahan email berhasil dikirim ke %s", email_lama)
		}
	}()

	log.Printf("[INFO] Email berhasil diubah untuk pengguna ID %d", id_pengguna)
	return &ResponseUbahEmail{
		Message: "Email berhasil diubah.",
	}

}
