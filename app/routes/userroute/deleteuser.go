package userroute

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	pengguna_service "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services"
)

func DeleteUserHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm
	ctx := r.Context()

	switch r.URL.Path {
	case "/user/komentar-barang/hapus":
		var data pengguna_service.PayloadHapusKomentarBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.HapusKomentarBarang(ctx, data, db)
	case "/user/keranjang-barang/hapus":
		var data pengguna_service.PayloadHapusDataKeranjangBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.HapusKeranjangBarang(ctx, data, db)
	default:
		hasil = &response.ResponseForm{
			Status:   http.StatusBadRequest,
			Services: "User Services",
			Payload:  "Gagal Coba Lagi Nanti",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
