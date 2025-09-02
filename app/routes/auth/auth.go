package auth

import (
	"encoding/json"
	"io"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service"
)

func HandleAuth(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/auth/user/registration" && r.Method == http.MethodPost {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data models.Pengguna
		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		hasil := service.UserRegistration(db, data.Username, data.Nama, data.Email, data.PasswordHash)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	if r.URL.Path == "/auth/user/login" && r.Method == http.MethodPost {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data models.Pengguna
		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		hasil := service.UserLogin(db, data.Email, data.PasswordHash)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	http.Error(w, "Endpoint tidak ditemukan", http.StatusNotFound)
}
