package seller

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_service "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services"
)

func PostSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm

	switch r.URL.Path {
	case "/seller/masukan_barang":
		var data seller_service.PayloadMasukanBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.MasukanBarang(db, data)
	case "/seller/tambah_kategori_barang":
		var data seller_service.PayloadTambahKategori
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.TambahKategoriBarang(db, data)
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
