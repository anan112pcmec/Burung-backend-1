package userroute

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	pengguna_service "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services"
)

func GetUserHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, rds *redis.Client) {
	var hasil *response.ResponseForm

	switch r.URL.Path {
	case "/user/barang-all":
		hasil = pengguna_service.AmbilRandomBarang(rds)
	case "/user/barang-spesified":
		jenis := r.URL.Query().Get("jenis")
		if jenis != "" {
			hasil = pengguna_service.AmbilBarangJenis(rds, jenis)
		} else {
			hasil = pengguna_service.AmbilRandomBarang(rds)
		}

	default:
		hasil = &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: "Seller Services",
			Payload:  "Gagal Coba Lagi Nanti",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
