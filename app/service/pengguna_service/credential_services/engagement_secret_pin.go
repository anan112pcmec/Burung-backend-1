package pengguna_credential_services

import (
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/credential_services/response_credential_pengguna"
)

func MembuatSecretPinPengguna(data PayloadMembuatPinPengguna, db *gorm.DB) *response.ResponseForm {
	services := "MembuatSecretPinPengguna"
	if data.IDPengguna == 0 {
		log.Println("[WARN] ID pengguna tidak ditemukan pada permintaan.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "ID pengguna tidak ditemukan.",
			},
		}
	}

	if data.Username == "" {
		log.Println("[WARN] Username tidak ditemukan pada permintaan.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Username tidak ditemukan.",
			},
		}
	}

	var user models.Pengguna
	if checkuser := db.Model(models.Pengguna{}).Select("pin_hash", "password_hash").Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Take(&user).Limit(1).Error; checkuser != nil {
		log.Printf("[WARN] Pengguna tidak ditemukan: %v", checkuser)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Pengguna tidak ditemukan.",
			},
		}
	}

	if check_pass := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password)); check_pass != nil {
		log.Println("[WARN] Password yang dimasukkan salah.")
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Password yang dimasukkan salah.",
			},
		}
	}

	hashed_pin, err := bcrypt.GenerateFromPassword([]byte(data.Pin), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Gagal mengenkripsi PIN: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Terjadi kesalahan pada server saat membuat PIN.",
			},
		}
	}

	if ganti_pin := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Update("pin_hash", hashed_pin).Error; ganti_pin != nil {
		log.Printf("[ERROR] Gagal menyimpan PIN ke database: %v", ganti_pin)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseMembuatPin{
				Message: "Terjadi kesalahan pada server saat menyimpan PIN.",
			},
		}
	}

	log.Println("[INFO] PIN berhasil dibuat.")
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseMembuatPin{
			Message: "PIN berhasil dibuat.",
		},
	}
}

func UpdateSecretPinPengguna(data PayloadUpdatePinPengguna, db *gorm.DB) *response.ResponseForm {
	services := "UpdateSecretPinPengguna"
	if data.IDPengguna == 0 {
		log.Println("[WARN] ID pengguna tidak ditemukan pada permintaan.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "ID pengguna tidak ditemukan.",
			},
		}
	}

	if data.Username == "" {
		log.Println("[WARN] Username tidak ditemukan pada permintaan.")
		return &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Username tidak ditemukan.",
			},
		}
	}

	var pin_lama string
	if checkuser := db.Model(models.Pengguna{}).Select("pin_hash").Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Limit(1).Take(&pin_lama).Error; checkuser != nil {
		log.Printf("[WARN] Pengguna tidak ditemukan: %v", checkuser)
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Pengguna tidak ditemukan.",
			},
		}
	}

	if check_pin := bcrypt.CompareHashAndPassword([]byte(pin_lama), []byte(data.PinLama)); check_pin != nil {
		log.Println("[WARN] PIN lama yang dimasukkan salah.")
		return &response.ResponseForm{
			Status:   http.StatusUnauthorized,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "PIN lama yang dimasukkan salah.",
			},
		}
	}

	hashed_pin, err := bcrypt.GenerateFromPassword([]byte(data.PinBaru), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Gagal mengenkripsi PIN baru: %v", err)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Terjadi kesalahan pada server saat membuat PIN baru.",
			},
		}
	}

	if err_ganti_pin := db.Model(models.Pengguna{}).Where(models.Pengguna{ID: data.IDPengguna, Username: data.Username}).Update("pin_hash", hashed_pin).Error; err_ganti_pin != nil {
		log.Printf("[ERROR] Gagal menyimpan PIN baru ke database: %v", err_ganti_pin)
		return &response.ResponseForm{
			Status:   http.StatusInternalServerError,
			Services: services,
			Payload: response_credential_pengguna.ResponseUpdatePin{
				Message: "Terjadi kesalahan pada server saat menyimpan PIN baru.",
			},
		}
	}

	log.Println("[INFO] PIN berhasil diubah.")
	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_credential_pengguna.ResponseUpdatePin{
			Message: "PIN berhasil diubah.",
		},
	}

}
