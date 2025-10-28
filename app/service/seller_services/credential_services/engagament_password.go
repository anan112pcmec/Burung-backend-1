package seller_credential_services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/credential_services/response_credential_seller"
)

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Pre Ubah Password Seller
// Berfungsi untuk mengirim kode otp ke gmail nantinya sebelum password benar benar diubah
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func PreUbahPasswordSeller(data PayloadPreUbahPasswordSeller, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "PreUbahPasswordSeller"

	seller, status := data.IdentitasSeller.Validating(db)
	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_seller.ResponsePreUbahPasswordSeller{
				Message: "Gagal Kredensial Seller Tidak Valid",
			},
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(seller.Password), []byte(data.PasswordLama)); err != nil {
		log.Println("[WARN] Password lama yang dimasukkan salah.")
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_seller.ResponsePreUbahPasswordSeller{
				Message: "Password lama yang dimasukkan salah.",
			},
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.PasswordBaru), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Gagal mengenkripsi password baru: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponsePreUbahPasswordSeller{
				Message: "Terjadi kesalahan pada server saat mengenkripsi password.",
			},
		}
	}

	go func() {
		otp := helper.GenerateOTP()
		key := fmt.Sprintf("seller_ubah_password_by_otp:%s", otp)

		var email string

		if err := db.Model(&models.Seller{}).
			Where(&models.Seller{
				ID: data.IdentitasSeller.IdSeller,
			}).
			Select("email").
			Take(&email).Error; err != nil {

			log.Printf("[ERROR] Gagal mengambil email seller ID %d: %v", data.IdentitasSeller.IdSeller, err)
			return
		}

		to := []string{email}
		subject := "Kode Mengubah Password Burung"
		message := fmt.Sprintf("Kode Anda: %s\nMasa berlaku 3 menit.", otp)

		if err := emailservices.SendMail(to, nil, subject, message); err != nil {
			log.Printf("[ERROR] Gagal mengirim email OTP ke %s: %v", email, err)
		} else {
			log.Printf("[INFO] Email OTP berhasil dikirim ke %s", email)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fields := map[string]interface{}{
			"id_seller":     data.IdentitasSeller,
			"username":      data.IdentitasSeller.Username,
			"password_baru": string(hashedPassword),
		}

		pipe := rds.TxPipeline()
		hset := pipe.HSet(ctx, key, fields)
		exp := pipe.Expire(ctx, key, 3*time.Minute)

		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("[ERROR] Redis pipeline gagal: %v", err)
		}

		_ = hset
		_ = exp
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponsePreUbahPasswordSeller{
			Message: "Kode OTP telah dikirim ke email Anda. Silakan cek email untuk melanjutkan proses ubah password.",
		},
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Fungsi Prosedur Validate Ubah Password Seller
// Berfungsi untuk memvalidasi dengan kode otp yang telah dikirimkan untuk mengubah password mereka
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ValidateUbahPasswordSeller(data PayloadValidateUbahPasswordSellerOTP, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordSeller"

	if data.IDSeller == 0 {
		log.Println("[WARN] ID seller tidak ditemukan pada permintaan validasi OTP.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
				Message: "ID seller tidak ditemukan.",
			},
		}
	}

	if data.OtpKeyValidateSeller == "" {
		log.Println("[WARN] OTP tidak ditemukan pada permintaan validasi OTP.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
				Message: "OTP tidak ditemukan.",
			},
		}
	}

	var id_seller int32
	if check_seller := db.Model(models.Seller{}).Select("id").Where(models.Seller{ID: data.IDSeller}).First(&id_seller).Error; check_seller != nil {
		log.Printf("[WARN] Seller tidak ditemukan untuk validasi OTP: %v", check_seller)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
				Message: "Seller tidak ditemukan.",
			},
		}
	}

	ctx := context.Background()
	key := fmt.Sprintf("seller_ubah_password_by_otp:%s", data.OtpKeyValidateSeller)

	result, err_rds := rds.HGetAll(ctx, key).Result()

	if err_rds != nil || len(result) == 0 {
		log.Printf("[WARN] OTP tidak valid atau sudah kadaluarsa: %v", err_rds)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
				Message: "OTP tidak valid atau sudah kadaluarsa.",
			},
		}
	}

	if err_change_pass := db.Model(models.Seller{}).Where(models.Seller{ID: data.IDSeller}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
		log.Printf("[ERROR] Gagal mengubah password seller ID %d: %v", data.IDSeller, err_change_pass)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
				Message: "Terjadi kesalahan pada server saat mengubah password.",
			},
		}
	}

	log.Printf("[INFO] Password seller ID %d berhasil diubah via OTP.", data.IDSeller)
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
			Message: "Password berhasil diubah.",
		},
	}
}
