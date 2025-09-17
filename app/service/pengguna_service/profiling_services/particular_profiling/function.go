package particular_profiling_pengguna

import (
	"context"
	"fmt"
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
		return &ResponseUbahUsername{
			Message: "Server Bermasalah",
		}
	}

	if countUsername == 0 {
		if err_seller := db.Model(&models.Pengguna{}).
			Where("id = ?", id_pengguna).
			Update("username", username).Error; err_seller == nil {
			return &ResponseUbahUsername{
				Message: "Berhasil",
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

		return &ResponseUbahUsername{
			Message: "Gagal, coba gunakan nama yang disarankan",
			Saran:   saran,
		}
	}

	return &ResponseUbahUsername{
		Message: "Gagal, Server sedang sibuk coba lagi nanti",
	}
}

func UbahNamaPengguna(id_pengguna int64, nama string, db *gorm.DB) *ResponseUbahNama {

	if err_db := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: id_pengguna}).Update("nama", nama).Error; err_db == nil {
		return &ResponseUbahNama{
			Message: "Berhasil",
		}
	}

	return &ResponseUbahNama{
		Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
	}
}

func UbahEmailSeller(ctx context.Context, id_pengguna int64, email string, db *gorm.DB) *ResponseUbahEmail {
	sudah_ada := ""
	if err_sama := db.Model(models.Pengguna{}).Select("email").Where(models.Pengguna{ID: id_pengguna, Email: email}).First(&sudah_ada).Error; err_sama != nil {
		fmt.Println("Gak Ada")
	}

	if sudah_ada != "" {
		return &ResponseUbahEmail{
			Message: "Gagal Mengubah Email, Coba Lagi Nanti",
		}
	}

	var email_lama string
	if err_email := db.Model(models.Pengguna{}).Select("email").Where(models.Pengguna{ID: id_pengguna}).First(&email_lama).Error; err_email != nil {
		return &ResponseUbahEmail{
			Message: "Gagal Mengubah Email, Coba Lagi Nanti",
		}
	}

	if err_ubah_email := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: id_pengguna}).Update("email", email).Error; err_ubah_email != nil {
		return &ResponseUbahEmail{
			Message: "Gagal Mengubah Email, Coba Lagi Nanti",
		}
	}

	go func() {
		to := []string{email_lama}
		cc := []string{}
		subject := "Pemberitahuan Pembaruan Gmail"
		message := fmt.Sprintf("Akun Burung Mu Telah Diubah Gmail nya Pada %s menjadi %s dan mulai sekarang semua pemberitahuan dan notif akan di berikan kepada email baru yang terkait", time.Time{}, email)
		err := emailservices.SendMail(to, cc, subject, message)
		if err != nil {
			fmt.Println("Gagal Kirim Pesan")
		}
	}()

	return &ResponseUbahEmail{
		Message: "Berhasil",
	}

}
