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
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/authservices"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/credential_services/response_credential_seller"
)

func PreUbahPasswordSeller(data PayloadPreUbahPasswordSeller, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "PreUbahPasswordSeller"

	if data.IDSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.Username == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var pass string
	if err := db.Model(&models.Seller{}).
		Select("password_hash").
		Where(&models.Seller{ID: data.IDSeller, Username: data.Username}).
		Pluck("password_hash", &pass).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload:  "Data Mu Tidak Valid",
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pass), []byte(data.PasswordLama)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload:  "Password Salah",
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.PasswordBaru), bcrypt.DefaultCost)
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
		}
	}

	go func() {
		otp := authservices.GenerateOTP()
		key := fmt.Sprintf("seller_ubah_password_by_otp:%s", otp)

		var email string

		if err := db.Model(&models.Seller{}).
			Select("password_hash").
			Where(&models.Seller{ID: data.IDSeller, Username: data.Username}).
			Pluck("email", &email).Error; err != nil {
			return
		}

		to := []string{email}
		subject := "Kode Mengubah Password Burung"
		message := fmt.Sprintf("Kode Anda: %s\nMasa berlaku 3 menit.", otp)

		if err := emailservices.SendMail(to, nil, subject, message); err != nil {
			fmt.Println("Gagal Kirim Email Untuk Otp:", otp)
		}

		log.Println("[TRACE] Email sent successfully")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fields := map[string]interface{}{
			"id_seller":     data.IDSeller,
			"username":      data.Username,
			"password_baru": string(hashedPassword),
		}

		pipe := rds.TxPipeline()
		hset := pipe.HSet(ctx, key, fields)
		exp := pipe.Expire(ctx, key, 3*time.Minute)

		if _, err := pipe.Exec(ctx); err != nil {
			log.Printf("[ERROR] Redis pipeline failed: %v\n", err)
		}

		_ = hset
		_ = exp
	}()

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponsePreUbahPasswordSeller{
			Message: "Berhasil",
		},
	}
}

func ValidateUbahPasswordSeller(data PayloadValidateUbahPasswordSellerOTP, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	services := "ValidateUbahPasswordSeller"

	if data.IDSeller == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if data.OtpKeyValidateSeller == "" {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	var id_seller int32
	if check_seller := db.Model(models.Seller{}).Select("id").Where(models.Seller{ID: data.IDSeller}).First(&id_seller).Error; check_seller != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	ctx := context.Background()
	key := fmt.Sprintf("seller_ubah_password_by_otp:%s", data.OtpKeyValidateSeller)

	result, err_rds := rds.HGetAll(ctx, key).Result()

	if err_rds != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if err_change_pass := db.Model(models.Seller{}).Where(models.Seller{ID: data.IDSeller}).Update("password_hash", string(result["password_baru"])).Error; err_change_pass != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_seller.ResponseValidateUbahPasswordSeller{
			Message: "Berhasil",
		},
	}
}
