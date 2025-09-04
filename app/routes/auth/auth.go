package auth

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/service/authservices"
)

type OTPkey struct {
	Value string `json:"otp_key"`
}

func HandleAuth(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/auth/user/registration":
		if r.Method == http.MethodPost {
			var data models.Pengguna
			if err := helper.DecodeJSONBody(r, &data); err != nil {
				http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			hasil := authservices.PreUserRegistration(db, data.Username, data.Nama, data.Email, data.PasswordHash)
			json.NewEncoder(w).Encode(hasil)
			return
		}

	case "/auth/user/login":
		if r.Method == http.MethodPost {
			var data models.Pengguna
			if err := helper.DecodeJSONBody(r, &data); err != nil {
				http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			hasil := authservices.UserLogin(db, data.Email, data.PasswordHash)
			json.NewEncoder(w).Encode(hasil)
			return
		}

	case "/auth/user/registration/validate":
		if r.Method == http.MethodPost {
			var data OTPkey
			if err := helper.DecodeJSONBody(r, &data); err != nil {
				http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			hasil := authservices.ValidateUserRegistration(db, data.Value)
			json.NewEncoder(w).Encode(hasil)
			return
		}

	case "/auth/seller/registration":
		if r.Method == http.MethodPost {
			var data models.Seller
			if err := helper.DecodeJSONBody(r, &data); err != nil {
				http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			hasil := authservices.PreSellerRegistration(db, data.Username, data.Nama, data.Email, data.Jenis, data.Norek, data.SellerDedication, data.Password)
			json.NewEncoder(w).Encode(hasil)
			return
		}

	case "/auth/seller/login":
		if r.Method == http.MethodPost {
			var data models.Seller
			if err := helper.DecodeJSONBody(r, &data); err != nil {
				http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			hasil := authservices.SellerLogin(db, data.Email, data.Password)
			json.NewEncoder(w).Encode(hasil)
			return
		}

	case "/auth/seller/registration/validate":
		if r.Method == http.MethodPost {
			var data OTPkey
			if err := helper.DecodeJSONBody(r, &data); err != nil {
				http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
				return
			}
			hasil := authservices.ValidateSellerRegistration(db, data.Value)
			json.NewEncoder(w).Encode(hasil)
			return
		}
	}

	http.Error(w, "Endpoint tidak ditemukan", http.StatusNotFound)
}
