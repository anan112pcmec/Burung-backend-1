package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/service/authservices"
)

type OTPkey struct {
	Value string `json:"otp_key"`
}

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

		hasil := authservices.PreUserRegistration(db, data.Username, data.Nama, data.Email, data.PasswordHash)

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

		hasil := authservices.UserLogin(db, data.Email, data.PasswordHash)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	if r.URL.Path == "/auth/user/registration/validate" && r.Method == http.MethodPost {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data OTPkey
		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		hasil := authservices.ValidateUserRegistration(db, data.Value)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	if r.URL.Path == "/auth/seller/registration" && r.Method == http.MethodPost {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data models.Seller

		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(map[string]interface{}{
			"username":          string(data.Username),
			"nama":              string(data.Nama),
			"email":             string(data.Email),
			"jenis":             string(data.Jenis),
			"norek":             string(data.Norek),
			"seller_dedication": string(data.SellerDedication),
			"password":          string(data.Password),
		})

		hasil := authservices.PreSellerRegistration(db, data.Username, data.Nama, data.Email, data.Jenis, data.Norek, data.SellerDedication, data.Password)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	if r.URL.Path == "/auth/seller/login" && r.Method == http.MethodPost {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data models.Seller
		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		hasil := authservices.SellerLogin(db, data.Email, data.Password)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	if r.URL.Path == "/auth/seller/registration/validate" && r.Method == http.MethodPost {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data OTPkey
		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		hasil := authservices.ValidateSellerRegistration(db, data.Value)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hasil)
		return
	}

	http.Error(w, "Endpoint tidak ditemukan", http.StatusNotFound)
}
