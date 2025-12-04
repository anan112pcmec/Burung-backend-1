package kurir

import (
	"encoding/json"
	"net/http"

	"github.com/anan112pcmec/Burung-backend-1/app/config"
	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	kurir_alamat_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/alamat_services"
	kurir_pengiriman_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/pengiriman_services"
	kurir_rekening_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/rekening_services"
)

func DeleteKurirHandler(db *config.InternalDBReadWriteSystem, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var hasil *response.ResponseForm
	switch r.URL.Path {
	case "/kurir/alamat/hapus-alamat":
		var data kurir_alamat_services.PayloadHapusAlamatKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_alamat_services.HapusAlamatKurir(ctx, data, db)
	case "/kurir/rekening/hapus-rekening":
		var data kurir_rekening_services.PayloadHapusRekeningKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_rekening_services.HapusRekeningKurir(ctx, data, db)
	case "/kurir/pengiriman/non-aktifkan-bid":
		var data kurir_pengiriman_services.PayloadNonaktifkanBidKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_pengiriman_services.NonaktifkanBidKurir(ctx, data, db)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
