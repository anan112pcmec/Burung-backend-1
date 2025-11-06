package kurir

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	kurir_alamat_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/alamat_services"
	kurir_informasi_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/informasi_services"
	kurir_pengiriman_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/pengiriman_services"
	kurir_rekening_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/rekening_services"
)

func PostKurirHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm
	switch r.URL.Path {
	case "/kurir/pengiriman/ambil-pengiriman":
		var data kurir_pengiriman_services.PayloadAmbilPengiriman
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_pengiriman_services.AmbilPengirimanKurir(data, db)
	case "/kurir/informasi/ajukan-informasi-kendaraan":
		var data kurir_informasi_services.PayloadInformasiDataKendaraan
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_informasi_services.AjukanInformasiKendaraan(data, db)

	case "/kurir/informasi/ajukan-informasi-kurir":
		var data kurir_informasi_services.PayloadInformasiDataKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_informasi_services.AjukanInformasiKurir(data, db)
	case "/kurir/alamat/masukan-alamat":
		var data kurir_alamat_services.PayloadMasukanAlamatKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_alamat_services.MasukanAlamatKurir(data, db)
	case "/kurir/rekening/masukan-rekening":
		var data kurir_rekening_services.PayloadMasukanRekeningKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_rekening_services.MasukanRekeningKurir(data, db)

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
