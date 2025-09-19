package pengguna_credential_services

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/credential_services/response_credential_pengguna"
)

func MembuatSecretPinPengguna(data PayloadMembuatPinPengguna, db *gorm.DB) *response.ResponseForm {
	services := "MembuatSecretPinPenggua"
	if data.IDPengguna == 0 {
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

	var user models.Pengguna
	if checkuser := db.Model(models.Pengguna{}).Select("pin_hash", "password_hash").Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Take(&user).Limit(1).Error; checkuser != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if check_pass := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password)); check_pass != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Gagal Password Salah",
			},
		}
	}

	hashed_pin, err := bcrypt.GenerateFromPassword([]byte(data.Pin), bcrypt.DefaultCost)
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Gagal MembuatPin, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	if ganti_pin := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Update("pin_hash", hashed_pin).Error; ganti_pin != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Gagal MembuatPin, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseMembuatPin{
			Message: "Berhasil",
		},
	}
}

func UpdateSecretPinPengguna(data PayloadUpdatePinPengguna, db *gorm.DB) *response.ResponseForm {
	services := "UpdatSecretPinPengguna"
	if data.IDPengguna == 0 {
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

	var pin_lama string
	if checkuser := db.Model(models.Pengguna{}).Select("pin_hash").Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Limit(1).Take(&pin_lama).Error; checkuser != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
		}
	}

	if check_pin := bcrypt.CompareHashAndPassword([]byte(pin_lama), []byte(data.PinLama)); check_pin != nil {
		return &response.ResponseForm{
			Status:   http.StatusNotAcceptable,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Gagal Pin Salah",
			},
		}
	}

	hashed_pin, err := bcrypt.GenerateFromPassword([]byte(data.PinBaru), bcrypt.DefaultCost)
	if err != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Gagal MembuatPin, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	if err_ganti_pin := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Update("pin_hash", hashed_pin).Error; err_ganti_pin != nil {
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Gagal MembuatPin, Server sedang sibuk coba lagi nanti",
			},
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseUpdatePin{
			Message: "Berhasil",
		},
	}

}
