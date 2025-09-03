package authservices

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	response_auth "github.com/anan112pcmec/Burung-backend-1/app/service/authservices/reponse_auth"
	"github.com/anan112pcmec/Burung-backend-1/app/service/emailservices"
)

func UserLogin(db *gorm.DB, email, password string) *response.ResponseForm {
	service := "UserLogin"
	var user models.Pengguna

	if err := db.Where(models.Pengguna{Email: email}).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: service,
				Payload: response_auth.LoginUserResp{
					Message: "Gagal Akun Belum Terdaftar, Coba Daftar kan dirimu dan bergabung bersama kami",
				},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: service,
			Payload: response_auth.LoginUserResp{
				Message: "Coba Login Nanti Lagi Server sedang sibuk",
			},
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: service,
			Payload: response_auth.LoginUserResp{
				Message: "Password Salah",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: service,
		Payload: response_auth.LoginUserResp{
			Status:  "Berhasil",
			Message: fmt.Sprintf("Kamu Berhasil Login Selamat datang %s", user.Nama),
			ID:      user.ID,
			LoginResponse: response_auth.LoginResponse{
				Nama:     user.Nama,
				Username: user.Username,
				Email:    user.Email,
			},
		},
	}
}

func SellerLogin(db *gorm.DB, email, password string) *response.ResponseForm {
	service := "SellerLogin"
	var seller models.Seller
	pass := "password"
	ID := "id"
	username := "username"
	emaild := "email"
	nama := "nama"

	if err := db.Where(models.Seller{Email: email}).Select(&ID, &nama, &username, &emaild, &pass).First(&seller).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: service,
				Payload: response_auth.LoginSellerResp{
					Message: "Gagal Akun Belum Terdaftar, Coba Daftar kan dirimu dan bergabung bersama kami",
				},
			}
		}

		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: service,
			Payload: response_auth.LoginSellerResp{
				Message: "Gagal Server Sedang Sibuk Coba Lagi Nanti",
			},
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(seller.Password), []byte(password)); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: service,
			Payload: response_auth.LoginSellerResp{
				Message: "Password Salah",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: service,
		Payload: response_auth.LoginSellerResp{
			Status:  "Berhasil",
			Message: fmt.Sprintf("Kamu Berhasil Login %s, Kembangkan Koneksimu Raih Keuntungan Disini", seller.Nama),
			ID:      seller.ID,
			LoginResponse: response_auth.LoginResponse{
				Nama:     seller.Nama,
				Username: seller.Username,
				Email:    seller.Email,
			},
		},
	}
}

func PreUserRegistration(db *gorm.DB, username, nama, email, password string) *response.ResponseForm {
	services := "PreUserRegistration"
	var user models.Pengguna

	if err := db.Where(models.Pengguna{Email: email}).Or(&models.Pengguna{Username: username}).Select("email", "username").First(&user).Error; err == nil {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_auth.PreRegistrationUserResp{
				Message: "Gagal Coba Ganti Username atau Gmail",
			},
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.PreRegistrationUserResp{
				Message: "Gagal Server Sedang Sibuk",
			},
		}
	}

	Otp := GenerateOTP()

	to := []string{email}
	cc := []string{}
	subject := "Kode OTP App Burung"
	message := fmt.Sprintf("Kode OTP Anda: %s\nMasa berlaku 3 menit.", Otp)

	if err := emailservices.SendMail(to, cc, subject, message); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.PreRegistrationUserResp{
				Message: "Gagal mengirim OTP, silakan coba lagi",
			},
		}
	}

	key := fmt.Sprintf("registration_user_pending:%s", Otp)

	fields := map[string]interface{}{
		"nama":          nama,
		"username":      username,
		"email":         email,
		"password_hash": password,
	}

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	for data, name := range fields {
		if err := rds.HSet(key, data, name).Err(); err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_auth.PreRegistrationUserResp{
					Message: "Kode Otp Sudah Terkirim Namun Server Sedang Terkendala, Coba Lagi Nanti dengan Kode Otp Lainnya",
				},
			}
		}
	}

	if err := rds.Expire(key, 3*time.Minute).Err(); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.PreRegistrationUserResp{
				Message: "Kode Otp Sudah Terkirim Namun Server Sedang Terkendala, Coba Lagi Nanti dengan Kode Otp Lainnya",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_auth.PreRegistrationUserResp{
			Status:  "Berhasil",
			Message: "Silahkan Masukan Kode OTP yang sudah di kirimkan ke Gmail Anda",
		},
	}
}

func ValidateUserRegistration(db *gorm.DB, OTPkey string) *response.ResponseForm {
	services := "ValidateUserRegistration"

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	key := fmt.Sprintf("registration_user_pending:%s", OTPkey)

	userData, err := rds.HGetAll(key).Result()
	if err != nil {
		fmt.Println("[ValidateUserRegistration] ERROR getting data from Redis:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.ValidateUserResp{
				Message: "Gagal Kode Sudah Expired, Coba Registrasi Ulang",
			},
		}
	}

	if len(userData) == 0 {
		fmt.Println("[ValidateUserRegistration] Key not found or expired")
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_auth.ValidateUserResp{
				Message: "Gagal Kode Sudah Expired, Coba Registrasi Ulang",
			},
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password_hash"]), bcrypt.DefaultCost)

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.ValidateUserResp{
				Message: "Gagal Server Sedang Sibuk, Coba Registrasi Ulang",
			},
		}
	}

	user := models.Pengguna{
		Nama:         userData["nama"],
		Username:     userData["username"],
		Email:        userData["email"],
		PasswordHash: string(hashedPassword),
	}

	if err := db.Unscoped().Create(&user).Error; err != nil {
		fmt.Println("[ValidateUserRegistration] ERROR saving to DB:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.ValidateUserResp{
				Message: "Gagal Server Sedang Sibuk, Coba Registrasi Ulang",
			},
		}
	}

	if err := rds.Del(key).Err(); err != nil {
		fmt.Println("[ValidateUserRegistration] WARNING deleting Redis key:", err)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_auth.ValidateUserResp{
			Status:  "Berhasil",
			Message: "Berhasil, Sekarang Kamu sudah memiliki akun di sistem Burung dan menjadi Bagian Dari Kami",
		},
	}
}

func PreSellerRegistration(db *gorm.DB, username, nama, email string, jenis models.JenisSeller, norek string, SellerDedication models.SellerType, password string) *response.ResponseForm {
	services := "PreSellerRegistration"
	var seller models.Seller

	if err := db.Unscoped().Where(models.Seller{Email: email}).Or(models.Seller{Username: username}).Select("email", "username").First(&seller).Error; err == nil {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload: response_auth.PreRegistrationSellerResp{
				Message: "Gagal Coba Ganti Username atau Gmail",
			},
		}
	}

	Otp := GenerateOTP()

	to := []string{email}
	cc := []string{}
	subject := "Kode OTP App Burung"
	message := fmt.Sprintf("Kode OTP Anda: %s\nMasa berlaku 3 menit.", Otp)

	if err := emailservices.SendMail(to, cc, subject, message); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.PreRegistrationSellerResp{
				Message: "Gagal Mengirim Kode OTP Coba Lagi Nanti",
			},
		}
	}

	key := fmt.Sprintf("registration_seller_pending:%s", Otp)

	fields := map[string]interface{}{
		"nama":              nama,
		"username":          username,
		"email":             email,
		"jenis":             jenis,
		"norek":             norek,
		"seller_dedication": SellerDedication,
		"password_hash":     password,
	}

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	for data, name := range fields {
		if err := rds.HSet(key, data, name).Err(); err != nil {
			return &response.ResponseForm{
				Status:   http.StatusInternalServerError,
				Services: services,
				Payload: response_auth.PreRegistrationSellerResp{
					Message: "Gagal Kirim Kode OTP Coba Lagi Nanti",
				},
			}
		}
	}

	if err := rds.Expire(key, 3*time.Minute).Err(); err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.PreRegistrationSellerResp{
				Message: "Gagal Kirim Kode OTP Coba Lagi Nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_auth.PreRegistrationSellerResp{
			Status:  "Berhasil",
			Message: "Silahkan Masukan Kode OTP yang sudah di kirimkan ke Gmail Anda",
		},
	}

}

func ValidateSellerRegistration(db *gorm.DB, OTPkey string) *response.ResponseForm {
	services := "ValidateSellerRegistration"

	rds := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})

	key := fmt.Sprintf("registration_seller_pending:%s", OTPkey)

	userData, err := rds.HGetAll(key).Result()
	if err != nil {
		fmt.Println("[ValidateUserRegistration] ERROR getting data from Redis:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.ValidateSellerResp{
				Message: "Gagal Kode Sudah Expired, Coba Registrasi Ulang",
			},
		}
	}

	if len(userData) == 0 {
		fmt.Println("[ValidateUserRegistration] Key not found or expired")
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_auth.ValidateSellerResp{
				Message: "Gagal Kode Sudah Expired, Coba Registrasi Ulang",
			},
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData["password_hash"]), bcrypt.DefaultCost)

	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.ValidateSellerResp{
				Message: "Gagal Server Sedang Sibuk, Coba Lagi Nanti",
			},
		}
	}

	seller := models.Seller{
		Nama:             userData["nama"],
		Username:         userData["username"],
		Email:            userData["email"],
		Jenis:            models.JenisSeller(userData["jenis"]),
		Norek:            userData["norek"],
		SellerDedication: models.SellerType(userData["seller_dedication"]),
		Password:         string(hashedPassword),
	}

	if err := db.Unscoped().Create(&seller).Error; err != nil {
		fmt.Println("[ValidateUserRegistration] ERROR saving to DB:", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_auth.ValidateSellerResp{
				Message: "Gagal Server Sedang Sibuk, Coba Lagi Nanti Ya",
			},
		}
	}

	if err := rds.Del(key).Err(); err != nil {
		fmt.Println("[ValidateUserRegistration] WARNING deleting Redis key:", err)
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_auth.ValidateSellerResp{
			Status:  "Berhasil",
			Message: "Berhasil, Akun mu sudah terdaftar dan Kamu Siap Berjualan Bersama Kami",
		},
	}
}
