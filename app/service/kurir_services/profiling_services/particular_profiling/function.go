package particular_profiling_kurir

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
)

func UbahNama(id_kurir int64, username_kurir, nama string, db *gorm.DB) ResponseUbahNama {
	if nama == "" {
		return ResponseUbahNama{
			Message: "Gagal",
		}
	}

	if err_ubah_nama := db.Model(models.Kurir{}).Where(models.Kurir{
		ID:       id_kurir,
		Username: username_kurir,
	}).Limit(1).Update("nama", nama).Error; err_ubah_nama != nil {
		return ResponseUbahNama{
			Message: "Gagal",
		}
	}

	return ResponseUbahNama{
		Message: "Berhasil",
	}
}

func UbahUsernameKurir(db *gorm.DB, id_kurir int64, username string) ResponseUbahUsername {
	if id_kurir == 0 {
		return ResponseUbahUsername{
			Message: "Gagal, ID Kurir Tidak Valid",
		}
	}

	var countUsername int64
	saran := make([]string, 0, 3)

	if err := db.Model(&models.Kurir{}).
		Where("username = ?", username).
		Count(&countUsername).Error; err != nil {
		return ResponseUbahUsername{
			Message: "Server Bermasalah",
		}
	}

	if countUsername == 0 {
		if err_update := db.Model(&models.Kurir{}).
			Where("id = ?", id_kurir).
			Update("username", username).Error; err_update == nil {

			return ResponseUbahUsername{
				Message: "Berhasil",
			}
		}
	}

	// jika username sudah ada, buat saran
	if countUsername > 0 {
		for len(saran) < 4 {
			usernameBaru := username + helper.GenerateRandomDigits()
			var tmp int64

			if err := db.Model(&models.Kurir{}).
				Where("username = ?", usernameBaru).
				Count(&tmp).Error; err != nil {
				continue
			}

			if tmp == 0 {
				duplicate := false
				for _, v := range saran {
					if v == usernameBaru {
						duplicate = true
						break
					}
				}
				if !duplicate {
					saran = append(saran, usernameBaru)
				}
			}
		}

		return ResponseUbahUsername{
			Message:       "Gagal, coba gunakan nama yang disarankan",
			SaranUsername: saran,
		}
	}

	return ResponseUbahUsername{
		Message: "Gagal, Server sedang sibuk coba lagi nanti",
	}
}

func UbahEmail(id_kurir int64, username, email string, db *gorm.DB) ResponseUbahGmail {
	if id_kurir == 0 && username == "" {
		return ResponseUbahGmail{
			Message: "Gagal",
		}
	}

	var emailnya string = ""
	if err_ambil_email := db.Model(models.Kurir{}).Select("email").Where(models.Kurir{
		ID:       id_kurir,
		Username: username,
	}).Limit(1).Take(&emailnya).Error; err_ambil_email != nil {
		return ResponseUbahGmail{
			Message: "Gagal",
		}
	}

	if err_ubah_email := db.Model(models.Kurir{}).Where(models.Kurir{ID: id_kurir, Username: username}).Limit(1).Update("email", email).Error; err_ubah_email != nil {
		return ResponseUbahGmail{
			Message: "Gagal",
		}
	}

	go func() {
		to := [...]string{
			emailnya,
		}
		cc := []string{}
		subject := "Pemberitahuan Pembaruan Gmail"
		message := fmt.Sprintf("Akun Burung Mu Telah Diubah Gmail nya Pada %s menjadi %s dan mulai sekarang semua pemberitahuan dan notif akan di berikan kepada email baru yang terkait", time.Time{}, email)
		err := emailservices.SendMail(to[:], cc, subject, message)
		if err != nil {
			fmt.Println("Gagal Kirim Pesan")
		}
	}()

	return ResponseUbahGmail{
		Message: "Berhasil",
	}

}

func UbahDeskripsi(id_kurir int64, username, deskripsi string, db *gorm.DB) *ResponseUbahDeskripsi {
	if username == "" && id_kurir == 0 {
		return &ResponseUbahDeskripsi{
			Message: "Gagal",
		}
	}

	if err_ubah_deskripsi := db.Model(models.Kurir{}).Where(models.Kurir{
		ID:       id_kurir,
		Username: username,
	}).Limit(1).Update("deskripsi", deskripsi).Error; err_ubah_deskripsi != nil {
		return &ResponseUbahDeskripsi{
			Message: "Gagal",
		}
	}

	return &ResponseUbahDeskripsi{
		Message: "Berhasil",
	}
}
