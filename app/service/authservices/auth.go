package authservices

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
func UserLogin(db *gorm.DB, email, password string
func UserLogin(db *gorm.DB, email, password string

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"

func UserLogin(db *gorm.DB, email, password string) *response.ResponseForm {
	service := "UserLogin"
	var user models.Pengguna

	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &response.ResponseForm{
				Status:   http.StatusNotFound, // 404
				Services: service,
				Payload:  nil,
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: service,
			Payload:  nil,
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: service,
			Payload:  nil,
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: service,
		Payload: response.LoginUserResp{
			ID: user.ID,
			LoginResponse: response.LoginResponse{
				Nama:     user.Nama,
				Username: user.Username,
				Email:    user.Email,
			},
		},
	}
}

func PreUserRegistration(db *gorm.DB, username, nama, email, password string) *response.ResponseForm {
	services := "PreUserRegistration"
	var user models.Pengguna

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	if err := db.Where("email = ? OR username = ?", email, username).First(&user).Error; err == nil {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload:  "Email atau username sudah terdaftar",
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Database error",
		}
	}

	Otp := GenerateOTP()

	user = models.Pengguna{
		Nama:         nama,
		Username:     username,
		Email:        email,
		PasswordHash: password,
	}

	to := []string{email}
	cc := []string{}
	subject := "Kode OTP App Burung"
	message := fmt.Sprintf("Kode OTP Anda: %s\nMasa berlaku 3 menit.", Otp)

	if err := emailservices.SendMail(to, cc, subject, message); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Gagal mengirim OTP, silakan coba lagi",
		}
	}

	// simpan di Redis Hash
	key := fmt.Sprintf("registration_pending:%s", Otp)

	fields := map[string]interface{}{
		"nama":          user.Nama,
		"username":      user.Username,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
	}

	for data, name := range fields {
		if err := rds.HSet(key, data, name).Err(); err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload:  "Gagal simpan data ke Redis",
			}
		}
	}

	if err := rds.Expire(key, 3*time.Minute).Err(); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Gagal set TTL Redis",
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  "Data pending registration tersimpan di Redis, OTP telah dikirim ke email",
	}
}

func ValidateUserRegistration(db *gorm.DB, OTPkey string) *response.ResponseForm {
	services := "ValidateUserRegistration"

	fmt.Println("Validate Registration jalan")

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	key := fmt.Sprintf("registration_pending:%s", OTPkey)
	fmt.Println("[ValidateUserRegistration] Checking Redis key:", key)

	userData, err := rds.HGetAll(key).Result()
	if err != nil {
		fmt.Println("[ValidateUserRegistration] ERROR getting data from Redis:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Gagal mengambil data dari Redis",
		}
	}

	if len(userData) == 0 {
		fmt.Println("[ValidateUserRegistration] Key not found or expired")
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "OTP tidak valid atau sudah kadaluarsa",
		}
	}

	user := models.Pengguna{
		Nama:         userData["nama"],
		Username:     userData["username"],
		Email:        userData["email"],
		PasswordHash: userData["password_hash"],
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Println("[ValidateUserRegistration] ERROR saving to DB:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  "Gagal menyimpan data user ke database",
		}
	}

	if err := rds.Del(key).Err(); err != nil {
		fmt.Println("[ValidateUserRegistration] WARNING deleting Redis key:", err)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  "User berhasil terdaftar",
	}
}
