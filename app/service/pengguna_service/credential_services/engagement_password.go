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
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/authservices"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/credential_services/response_credential_pengguna"
)

func PreUbahPasswordPengguna(data PayloadPreUbahPasswordPengguna, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "PreUbahPasswordPengguna"

	if data.FaktorKedua != "OTP" && data.FaktorKedua != "PIN" {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
		}
	}

	if data.IDPengguna == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
		}
	}

	var user models.Pengguna
	if password_check := db.Select("password_hash", "email", "pin_hash").
		Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).
		Limit(1).Take(&user).Error; password_check != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.PasswordSebelum)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
		}
	} else {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.PasswordBaru), bcrypt.DefaultCost)
		if err != nil {
			return &response.ResponseForm{
				Status:   http.StatusUnauthorized,
				Services: services,
			}
		}
		go func() {
			if data.FaktorKedua == "OTP" {

				otp := authservices.GenerateOTP()
				key := fmt.Sprintf("user_ubah_password_by_otp:%s", otp)

				to := []string{user.Email}
				subject := "Kode Mengubah Password Burung"
				message := fmt.Sprintf("Kode Anda: %s\nMasa berlaku 3 menit.", otp)

				if err := emailservices.SendMail(to, nil, subject, message); err != nil {
					fmt.Println("Gagal Kirim Email Untuk Otp:", otp)
				}

				log.Println("[TRACE] Email sent successfully")

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

				res, err := pipe.Exec(ctx)
				if err != nil {
					log.Printf("[ERROR] Redis pipeline failed: %v\n", err)
				}

				_ = hset
				_ = exp
				_ = res
			} else {
				key := fmt.Sprintf("user_ubah_password_by_pin:%v", data.IDPengguna)

				to := []string{user.Email}
				subject := "Kode Mengubah Password Burung"
				message := fmt.Sprintf("Kamu Mengubah Password Akun Burung pada %s, menggunakan faktor pin", time.Now())

				if err := emailservices.SendMail(to, nil, subject, message); err != nil {
					fmt.Println("Gagal Kirim Notifikasi Email")
				}

				log.Println("[TRACE] Email sent successfully")

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

				res, err := pipe.Exec(ctx)
				if err != nil {
					log.Printf("[ERROR] Redis pipeline failed: %v\n", err)
				}

				_ = hset
				_ = exp
				_ = res
			}
		}()
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponsePreUbahPassword{
			Message: fmt.Sprintf("Berhasil Silahkan Masukan Kredensial Yang diminta Selanjutnya Yakni: %s", data.FaktorKedua),
		},
	}
}

func ValidateUbahPasswordPenggunaViaOtp(data PayloadValidateOTPPasswordPengguna, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordPenggunaViaOtp"

	var id_user int64
	if check_user := db.Model(models.Pengguna{}).Select("id").Where(models.Pengguna{ID: data.IDPengguna}).First(&id_user).Error; check_user != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	ctx := context.Background()
	key := fmt.Sprintf("user_ubah_password_by_otp:%s", data.OtpKey)

	result, err_rds := rds.HGetAll(ctx, key).Result()

	if err_rds != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if err_change_pass := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseValidatePassword{
			Message: "Berhasil",
		},
	}

}

func ValidateUbahPasswordPenggunaViaPin(data PayloadValidatePinPasswordPengguna, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordPenggunaViaPin"

	var pin_user string
	if check_pin := db.Model(models.Pengguna{}).Select("pin_hash").Where(models.Pengguna{ID: data.IDPengguna}).First(&pin_user).Error; check_pin != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pin_user), []byte(data.Pin)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
		}
	} else {
		ctx := context.Background()
		key := fmt.Sprintf("user_ubah_password_by_pin:%v", data.IDPengguna)
		result, err_validate := rds.HGetAll(ctx, key).Result()
		if err_validate != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
			}
		}

		if err_change_pass := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseValidatePassword{
			Message: "Berhasil",
		},
	}
}
