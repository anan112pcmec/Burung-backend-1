package kurir

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	kurir_informasi_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/informasi_services"
	kurir_pengiriman_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/pengiriman_services"
	kurir_profiling_service "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/profiling_services"
)

func PatchKurirHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm
	switch r.URL.Path {
	case "/kurir/profiling/personal-update":
		var data kurir_profiling_service.PayloadPersonalProfilingKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_profiling_service.PersonalProfilingKurir(data, db)
	case "/kurir/informasi/edit-informasi-kendaraan":
		var data kurir_informasi_services.PayloadEditInformasiDataKendaraan
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_informasi_services.EditInformasiKendaraan(data, db)
	case "/kurir/informasi/edit-informasi-kurir":
		var data kurir_informasi_services.PayloadEditInformasiDataKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_informasi_services.EditInformasiKurir(data, db)
	case "/kurir/pengiriman/update-pengiriman":
		var data kurir_pengiriman_services.PayloadUpdatePengiriman
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_pengiriman_services.UpdatePengirimanKurir(data, db)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
