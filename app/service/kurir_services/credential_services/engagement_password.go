package kurir_credential_services

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
	response_credential_kurir "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/credential_services/response_credential_services"
)

func PreUbahPasswordKurir(data PayloadPreUbahPassword, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "PreUbahPasswordKurir"

	kurir, status := data.DataIdentitas.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_kurir.ResponsePreUbahPassword{
				Message: "Gagal, identitas kurir tidak valid",
			},
		}
	}

	if data.PasswordBaru == "" && data.PasswordLama == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_kurir.ResponsePreUbahPassword{
				Message: "Gagal, Isi Password Yang Benar",
			},
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(kurir.PasswordHash), []byte(data.PasswordLama)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
			Payload: response_credential_kurir.ResponsePreUbahPassword{
				Message: "Gagal, Password Lama Mu Salah",
			},
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.PasswordBaru), bcrypt.DefaultCost)
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_kurir.ResponsePreUbahPassword{
				Message: "Gagal, server sedang sibuk coba lagi lain waktu",
			},
		}
	}

	go func() {
		otp := authservices.GenerateOTP()
		key := fmt.Sprintf("kurir_ubah_password_by_otp:%s", otp)

		to := []string{data.DataIdentitas.EmailKurir}
		subject := "Kode Mengubah Password Burung"
		message := fmt.Sprintf("Kode Anda: %s\nMasa berlaku 3 menit.", otp)

		if err := emailservices.SendMail(to, nil, subject, message); err != nil {
			fmt.Println("Gagal Kirim Email Untuk Otp:", otp)
		}

		log.Println("[TRACE] Email sent successfully")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fields := map[string]interface{}{
			"id_kurir":      data.DataIdentitas.IdKurir,
			"username":      data.DataIdentitas.UsernameKurir,
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

	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_kurir.ResponsePreUbahPassword{
			Message: fmt.Sprintf("Berhasil, Selanjutnya silahkan input kode otp yang dikirim ke email anda: %s", data.DataIdentitas.EmailKurir),
		},
	}

}

func ValidateUbahPasswordKurir(data PayloadValidateUbahPassword, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordKurir"

	_, status := data.DataIdentitas.Validating(db)

	if !status {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_kurir.ResponseValidateUbahPassword{
				Message: "Gagal, identitas kurir tidak valid",
			},
		}
	}

	ctx := context.Background()
	key := fmt.Sprintf("kurir_ubah_password_by_otp:%s", data.OtpKey)

	result, err_rds := rds.HGetAll(ctx, key).Result()

	if err_rds != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if err_change_pass := db.Model(models.Kurir{}).Where(models.Kurir{ID: data.DataIdentitas.IdKurir}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_kurir.ResponseValidateUbahPassword{
			Message: "Berhasil",
		},
	}
}
