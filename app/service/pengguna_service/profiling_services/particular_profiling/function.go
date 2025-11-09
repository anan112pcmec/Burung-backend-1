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

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur mengubah username pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahUsernamePengguna(ctx context.Context, db *gorm.DB, id_pengguna int64, username string) *ResponseUbahUsername {
	if username == "not" {
		return &ResponseUbahUsername{
			Status: false,
		}
	}

	var countUsername int64

	if err := db.WithContext(ctx).Model(&models.Pengguna{}).
		Where(&models.Pengguna{Username: username}).
		Count(&countUsername).Error; err != nil {
		log.Printf("[ERROR] Gagal memeriksa username: %v", err)
		return &ResponseUbahUsername{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			Status:  true,
		}
	}

	// Jika username belum digunakan, langsung ubah
	if countUsername == 0 {
		if err_update := db.WithContext(ctx).Model(&models.Pengguna{}).
			Where(&models.Pengguna{ID: id_pengguna}).
			Update("username", username).Error; err_update == nil {
			log.Printf("[INFO] Username berhasil diubah untuk pengguna ID %d", id_pengguna)
			return &ResponseUbahUsername{
				Message: "Username berhasil diubah.",
				Status:  true,
			}
		}

		log.Printf("[ERROR] Gagal mengubah username untuk pengguna ID %d", id_pengguna)
		return &ResponseUbahUsername{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			Status:  true,
		}
	}

	// Jika username sudah digunakan, buat saran alternatif
	if countUsername > 0 {
		saran := make([]string, 0, 4)
		for len(saran) < 4 {
			usernameBaru := username + helper.GenerateRandomDigits()
			var tmp int64

			if err := db.WithContext(ctx).Model(&models.Pengguna{}).
				Where(&models.Pengguna{Username: usernameBaru}).
				Count(&tmp).Error; err != nil {
				log.Printf("[WARN] Gagal memeriksa ketersediaan saran username: %v", err)
				continue
			}

			if tmp == 0 {
				duplikat := false
				for _, v := range saran {
					if v == usernameBaru {
						duplikat = true
						break
					}
				}
				if !duplikat {
					saran = append(saran, usernameBaru)
				}
			}
		}

		log.Printf("[WARN] Username '%s' sudah digunakan. Menyediakan saran.", username)
		return &ResponseUbahUsername{
			Message: "Username sudah digunakan. Silakan pilih salah satu saran berikut.",
			Saran:   saran,
			Status:  true,
		}
	}

	// Fallback jika kondisi tak terduga
	log.Printf("[ERROR] Kondisi tak terduga saat mengubah username ID %d", id_pengguna)
	return &ResponseUbahUsername{
		Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
		Status:  true,
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur mengubah nama pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahNamaPengguna(ctx context.Context, id_pengguna int64, nama string, db *gorm.DB) *ResponseUbahNama {
	if nama == "not" {
		return &ResponseUbahNama{
			Status: false,
		}
	}

	if err_db := db.WithContext(ctx).Model(&models.Pengguna{}).Where(&models.Pengguna{ID: id_pengguna}).Update("nama", nama).Error; err_db == nil {
		log.Printf("[INFO] Nama berhasil diubah untuk pengguna ID %d", id_pengguna)
		return &ResponseUbahNama{
			Message: "Nama berhasil diubah.",
			Status:  true,
		}
	}

	log.Printf("[ERROR] Gagal mengubah nama untuk pengguna ID %d", id_pengguna)
	return &ResponseUbahNama{
		Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
		Status:  true,
	}
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur mengubah email pengguna
// ////////////////////////////////////////////////////////////////////////////////////////////////////////////

func UbahEmailPengguna(ctx context.Context, id_pengguna int64, email string, db *gorm.DB) *ResponseUbahEmail {

	if email == "not" {
		return &ResponseUbahEmail{
			Status: false,
		}
	}

	// cek apakah email yang dimasukkan sama dengan yang sudah tersimpan untuk pengguna ini
	var id_data_lama int64 = 0
	if err := db.WithContext(ctx).Model(&models.Pengguna{}).
		Where(&models.Pengguna{ID: id_pengguna, Email: email}).
		Limit(1).Scan(&id_data_lama).Error; err != nil {
		// tidak fatal, hanya log info — lanjutkan
		log.Printf("[INFO] Gagal mengecek kesamaan email untuk pengguna ID %d: %v", id_pengguna, err)
	}

	if id_data_lama > 0 {
		log.Printf("[WARN] Email yang dimasukkan sama dengan email lama untuk pengguna ID %d", id_pengguna)
		return &ResponseUbahEmail{
			Message: "Email yang dimasukkan sama dengan email sebelumnya.",
			Status:  true,
		}
	}

	// ambil email lama
	var email_lama string = ""
	if err := db.WithContext(ctx).Model(&models.Pengguna{}).
		Select("email").
		Where(&models.Pengguna{ID: id_pengguna}).
		Limit(1).Scan(&email_lama).Error; err != nil {
		log.Printf("[ERROR] Gagal mendapatkan email lama pengguna ID %d: %v", id_pengguna, err)
		return &ResponseUbahEmail{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			Status:  true,
		}
	}

	// update email
	if err := db.WithContext(ctx).Model(&models.Pengguna{}).
		Where(&models.Pengguna{ID: id_pengguna}).
		Update("email", email).Error; err != nil {
		log.Printf("[ERROR] Gagal mengubah email untuk pengguna ID %d: %v", id_pengguna, err)
		return &ResponseUbahEmail{
			Message: "Terjadi kesalahan pada server. Silakan coba lagi nanti.",
			Status:  true,
		}
	}

	to := []string{email_lama}
	cc := []string{}
	subject := "Pemberitahuan Pembaruan Email"
	message := fmt.Sprintf("Akun Burung Anda telah diubah email-nya pada %s menjadi %s. Mulai sekarang semua pemberitahuan akan dikirim ke email baru tersebut.", time.Now().Format("02-01-2006 15:04:05"), email)
	if err := emailservices.SendMail(to, cc, subject, message); err != nil {
		log.Printf("[ERROR] Gagal mengirim email notifikasi perubahan email ke %s: %v", email_lama, err)
	} else {
		log.Printf("[INFO] Notifikasi perubahan email berhasil dikirim ke %s", email_lama)
	}

	log.Printf("[INFO] Email berhasil diubah untuk pengguna ID %d", id_pengguna)
	return &ResponseUbahEmail{
		Message: "Email berhasil diubah.",
		Status:  true,
	}
}
