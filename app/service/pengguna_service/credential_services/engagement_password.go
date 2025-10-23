package pengguna_credential_services

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
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/credential_services/response_credential_pengguna"
)

func PreUbahPasswordPengguna(data PayloadPreUbahPasswordPengguna, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "PreUbahPasswordPengguna"

	if data.FaktorKedua != "OTP" && data.FaktorKedua != "PIN" {
		log.Printf("[WARN] Faktor kedua tidak valid: %s", data.FaktorKedua)
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_pengguna.ResponsePreUbahPassword{
				Message: "Faktor kedua tidak valid. Gunakan OTP atau PIN.",
			},
		}
	}

	if data.IDPengguna == 0 {
		log.Println("[WARN] ID pengguna tidak ditemukan pada permintaan.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_pengguna.ResponsePreUbahPassword{
				Message: "ID pengguna tidak ditemukan.",
			},
		}
	}

	var user models.Pengguna
	if password_check := db.Select("password_hash", "email", "pin_hash").
		Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).
		Limit(1).Take(&user).Error; password_check != nil {
		log.Printf("[WARN] Pengguna tidak ditemukan: %v", password_check)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_pengguna.ResponsePreUbahPassword{
				Message: "Pengguna tidak ditemukan atau kredensial salah.",
			},
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.PasswordSebelum)); err != nil {
		log.Println("[WARN] Password lama tidak sesuai.")
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_pengguna.ResponsePreUbahPassword{
				Message: "Password lama tidak sesuai.",
			},
		}
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.PasswordBaru), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[ERROR] Gagal mengenkripsi password baru: %v", err)
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_credential_pengguna.ResponsePreUbahPassword{
					Message: "Terjadi kesalahan pada server saat mengenkripsi password.",
				},
			}
		}
		go func() {
			if data.FaktorKedua == "OTP" {

				otp := helper.GenerateOTP()
				key := fmt.Sprintf("user_ubah_password_by_otp:%s", otp)

				to := []string{user.Email}
				subject := "Kode Mengubah Password Burung"
				message := fmt.Sprintf("Kode Anda: %s\nMasa berlaku 3 menit.", otp)

				if err := emailservices.SendMail(to, nil, subject, message); err != nil {
					log.Printf("[ERROR] Gagal mengirim email OTP: %v", err)
				} else {
					log.Println("[INFO] Email OTP berhasil dikirim.")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				fields := map[string]interface{}{
					"id_user":       data.IDPengguna,
					"username":      data.Username,
					"password_baru": string(hashedPassword),
				}

				pipe := rds.TxPipeline()
				hset := pipe.HSet(ctx, key, fields)
				exp := pipe.Expire(ctx, key, 3*time.Minute)

				_, err := pipe.Exec(ctx)
				if err != nil {
					log.Printf("[ERROR] Gagal menyimpan OTP ke Redis: %v", err)
				}

				_ = hset
				_ = exp
			} else {
				key := fmt.Sprintf("user_ubah_password_by_pin:%v", data.IDPengguna)

				to := []string{user.Email}
				subject := "Kode Mengubah Password Burung"
				message := fmt.Sprintf("Anda mengubah password akun Burung pada %s menggunakan faktor PIN.", time.Now().Format("02-01-2006 15:04:05"))

				if err := emailservices.SendMail(to, nil, subject, message); err != nil {
					log.Printf("[ERROR] Gagal mengirim notifikasi email PIN: %v", err)
				} else {
					log.Println("[INFO] Email notifikasi PIN berhasil dikirim.")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				fields := map[string]interface{}{
					"id_user":       data.IDPengguna,
					"username":      data.Username,
					"password_baru": string(hashedPassword),
				}

				pipe := rds.TxPipeline()
				hset := pipe.HSet(ctx, key, fields)
				exp := pipe.Expire(ctx, key, 3*time.Minute)

				_, err := pipe.Exec(ctx)
				if err != nil {
					log.Printf("[ERROR] Gagal menyimpan data PIN ke Redis: %v", err)
				}

				_ = hset
				_ = exp
			}
		}()
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponsePreUbahPassword{
			Message: fmt.Sprintf("Berhasil, silakan masukkan kredensial berikutnya: %s.", data.FaktorKedua),
		},
	}
}

func ValidateUbahPasswordPenggunaViaOtp(data PayloadValidateOTPPasswordPengguna, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordPenggunaViaOtp"

	var id_user int64
	if check_user := db.Model(models.Pengguna{}).Select("id").Where(models.Pengguna{ID: data.IDPengguna}).First(&id_user).Error; check_user != nil {
		log.Printf("[WARN] Pengguna tidak ditemukan untuk validasi OTP: %v", check_user)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_pengguna.ResponseValidatePassword{
				Message: "Pengguna tidak ditemukan.",
			},
		}
	}

	ctx := context.Background()
	key := fmt.Sprintf("user_ubah_password_by_otp:%s", data.OtpKey)

	result, err_rds := rds.HGetAll(ctx, key).Result()

	if err_rds != nil || len(result) == 0 {
		log.Printf("[WARN] OTP tidak valid atau sudah kadaluarsa: %v", err_rds)
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_pengguna.ResponseValidatePassword{
				Message: "OTP tidak valid atau sudah kadaluarsa.",
			},
		}
	}

	if err_change_pass := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
		log.Printf("[ERROR] Gagal mengubah password via OTP: %v", err_change_pass)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseValidatePassword{
				Message: "Terjadi kesalahan pada server saat mengubah password.",
			},
		}
	}

	log.Println("[INFO] Password berhasil diubah via OTP.")
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseValidatePassword{
			Message: "Password berhasil diubah.",
		},
	}

}

func ValidateUbahPasswordPenggunaViaPin(data PayloadValidatePinPasswordPengguna, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordPenggunaViaPin"

	var pin_user string
	if check_pin := db.Model(models.Pengguna{}).Select("pin_hash").Where(models.Pengguna{ID: data.IDPengguna}).First(&pin_user).Error; check_pin != nil {
		log.Printf("[WARN] PIN pengguna tidak ditemukan: %v", check_pin)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_pengguna.ResponseValidatePassword{
				Message: "PIN pengguna tidak ditemukan.",
			},
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pin_user), []byte(data.Pin)); err != nil {
		log.Println("[WARN] PIN yang dimasukkan salah.")
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_pengguna.ResponseValidatePassword{
				Message: "PIN yang dimasukkan salah.",
			},
		}
	} else {
		ctx := context.Background()
		key := fmt.Sprintf("user_ubah_password_by_pin:%v", data.IDPengguna)
		result, err_validate := rds.HGetAll(ctx, key).Result()
		if err_validate != nil || len(result) == 0 {
			log.Printf("[WARN] Data perubahan password via PIN tidak ditemukan atau sudah kadaluarsa: %v", err_validate)
			return &response.ResponseForm{
				Status:   http.StatusUnauthorized,
				Services: services,
				Payload: response_credential_pengguna.ResponseValidatePassword{
					Message: "Data perubahan password via PIN tidak ditemukan atau sudah kadaluarsa.",
				},
			}
		}

		if err_change_pass := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
			log.Printf("[ERROR] Gagal mengubah password via PIN: %v", err_change_pass)
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_credential_pengguna.ResponseValidatePassword{
					Message: "Terjadi kesalahan pada server saat mengubah password.",
				},
			}
		}
	}

	log.Println("[INFO] Password berhasil diubah via PIN.")
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseValidatePassword{
			Message: "Password berhasil diubah.",
		},
	}
}
