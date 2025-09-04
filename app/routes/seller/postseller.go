package seller

import (
	"encoding/json"
	"io"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_service "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services"
)

func PostSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm

	if r.URL.Path == "/seller/masukan-barang" {
		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Gagal membaca body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var data seller_service.PayloadMasukanBarang
		if err := json.Unmarshal(bb, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		hasil = seller_service.MasukanBarang(db, data)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
