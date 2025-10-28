package seller_particular_profiling

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Particular Ubah Username Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahUsernameSeller(id_seller int32, username string, db *gorm.DB) *ResponseUbahUsername {
	var countUsername int64
	saran := make([]string, 0, 3)

	if err := db.Model(&models.Seller{}).
		Where("username = ?", username).
		Count(&countUsername).Error; err != nil {
		return &ResponseUbahUsername{
			Message: "Server Bermasalah",
		}
	}

	if countUsername == 0 {
		if err_seller := db.Model(&models.Seller{}).
			Where("id = ?", id_seller).
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

			if err := db.Model(&models.Seller{}).
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
			_ = db.Create(&models.AktivitasSeller{
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

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Particular Ubah Nama Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahNamaSeller(id_seller int32, nama string, db *gorm.DB) *ResponseUbahNama {

	if err_db := db.Model(models.Seller{}).Where(models.Seller{ID: id_seller}).Update("nama", nama).Error; err_db == nil {
		return &ResponseUbahNama{
			Message: "Berhasil",
		}
	}

	return &ResponseUbahNama{
		Message: "Gagal, Server Sedang Sibuk Coba Lagi Nanti",
	}
}

func UbahEmailSeller(ctx context.Context, id_seller int32, email string, db *gorm.DB) *ResponseUbahEmail {
	sudah_ada := ""
	if err_sama := db.Model(models.Seller{}).Select("email").Where(models.Seller{ID: id_seller, Email: email}).First(&sudah_ada).Error; err_sama != nil {
		fmt.Println("Gak Ada")
	}

	if sudah_ada != "" {
		return &ResponseUbahEmail{
			Message: "Gagal Mengubah Email, Coba Lagi Nanti",
		}
	}

	var email_lama string
	if err_email := db.Model(models.Seller{}).Select("email").Where(models.Seller{ID: id_seller}).First(&email_lama).Error; err_email != nil {
		return &ResponseUbahEmail{
			Message: "Gagal Mengubah Email, Coba Lagi Nanti",
		}
	}

	if err_ubah_email := db.Model(models.Seller{}).Where(models.Seller{ID: id_seller}).Update("email", email).Error; err_ubah_email != nil {
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

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Particular Ubah Dedication Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahSellerDedicationSeller(id_seller int32, seller_dedication string, db *gorm.DB) {

}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Particular Ubah Jam Operasional Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahJamOperasionalSeller(id_seller int32, username string, jam_operasional string, db *gorm.DB) *ResponseUbahJamOperasional {
	if id_seller == 0 && username == "" {
		return &ResponseUbahJamOperasional{
			Message: "Gagal Id seller tidak Valid",
		}
	}

	if err_ubah := db.Model(models.Seller{}).Where(models.Seller{ID: id_seller, Username: username}).Update("jam_operasional", jam_operasional).Error; err_ubah != nil {
		return &ResponseUbahJamOperasional{
			Message: "Gagal Server Sedang sibuk, Coba Lagi lain waktu",
		}
	}

	return &ResponseUbahJamOperasional{
		Message: "Berhasil",
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Particular Ubah Punchline Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahPunchlineSeller(id_seller int32, username string, punchline string, db *gorm.DB) *ResponseUbahPunchline {
	if id_seller == 0 && username == "" {
		return &ResponseUbahPunchline{
			Message: "Gagal, Seller Tidak Valid",
		}
	}

	if err_ubah := db.Model(models.Seller{}).Where(models.Seller{ID: id_seller, Username: username}).Update("punchline", punchline).Error; err_ubah != nil {
		return &ResponseUbahPunchline{
			Message: "Gagal Mungkin Server Sedang sibuk",
		}
	}

	return &ResponseUbahPunchline{
		Message: "Berhasil",
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Particular Ubah Deskripsi Seller
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahDeskripsiSeller(id_seller int32, username string, deskripsi string, db *gorm.DB) *ResponseUbahDeskripsi {
	if id_seller == 0 && username == "" {
		return &ResponseUbahDeskripsi{
			Message: "Gagal, Seller kredensial tidak valid",
		}
	}

	if err_ubah := db.Model(models.Seller{}).Where(models.Seller{ID: id_seller, Username: username}).Update("deskripsi", deskripsi).Error; err_ubah != nil {
		return &ResponseUbahDeskripsi{
			Message: "Gagal Server Sedang Sibuk, Coba Lagi Lain Waktu",
		}
	}

	return &ResponseUbahDeskripsi{
		Message: "Berhasil",
	}
}
