package seller

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_service "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services"
)

func DeleteSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm

	switch r.URL.Path {
	case "/seller/hapus_barang":
		var data seller_service.PayloadHapusBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
		}
		hasil = seller_service.HapusBarang(db, data)
	case "/seller/hapus_kategori_barang":
		var data seller_service.PayloadHapusKategori
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
		}
		hasil = seller_service.HapusKategoriBarang(db, data)
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
