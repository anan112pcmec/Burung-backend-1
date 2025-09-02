package service

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"


)

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

func UserRegistration(db *gorm.DB, username, nama, email, password string) *response.ResponseForm {
	services := "registration user"
	var user models.Pengguna

	if err := db.Where("email = ? OR username = ?", email, username).First(&user).Error; err == nil {
		return &response.ResponseForm{
			Status:   http.StatusConflict,
			Services: services,
			Payload:  nil,
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  nil,
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  nil,
		}
	}

	user = models.Pengguna{
		Nama:         nama,
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := db.Create(&user).Error; err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload:  nil,
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  nil,
	}
}
