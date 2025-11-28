package kurir

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	kurir_alamat_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/alamat_services"
	kurir_credential_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/credential_services"
	kurir_informasi_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/informasi_services"
	kurir_pengiriman_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/pengiriman_services"
	kurir_profiling_service "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/profiling_services"
	kurir_rekening_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/rekening_services"
	kurir_social_media_services "github.com/anan112pcmec/Burung-backend-1/app/service/kurir_services/social_media_services"
)

func PatchKurirHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, rds *redis.Client) {
	var hasil *response.ResponseForm
	ctx := r.Context()
	switch r.URL.Path {
	case "/kurir/profiling/personal-update":
		var data kurir_profiling_service.PayloadPersonalProfilingKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_profiling_service.PersonalProfilingKurir(ctx, data, db)
	case "/kurir/profiling/general-update":
		var data kurir_profiling_service.PayloadGeneralProfiling
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_profiling_service.GeneralProfilingKurir(ctx, data, db)
	case "/kurir/informasi/edit-informasi-kendaraan":
		var data kurir_informasi_services.PayloadEditInformasiDataKendaraan
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_informasi_services.EditInformasiKendaraan(ctx, data, db)
	case "/kurir/informasi/edit-informasi-kurir":
		var data kurir_informasi_services.PayloadEditInformasiDataKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_informasi_services.EditInformasiKurir(ctx, data, db)
	case "/kurir/social-media/social-media-engage":
		var data kurir_social_media_services.PayloadEngageSocialMedia
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_social_media_services.EngagementSocialMediaKurir(ctx, data, db)
	case "/kurir/alamat/edit-alamat":
		var data kurir_alamat_services.PayloadEditAlamatKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_alamat_services.EditAlamatKurir(ctx, data, db)
	case "/kurir/rekening/edit-rekening":
		var data kurir_rekening_services.PayloadEditRekeningKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_rekening_services.EditRekeningKurir(ctx, data, db)

	case "/kurir/credential/preubah-pass":
		var data kurir_credential_services.PayloadPreUbahPassword
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_credential_services.PreUbahPasswordKurir(ctx, data, db, rds)
	case "/kurir/credential/validate-ubah-pass-otp":
		var data kurir_credential_services.PayloadValidateUbahPassword
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_credential_services.ValidateUbahPasswordKurir(ctx, data, db, rds)
	case "/kurir/pengiriman/aktifkan-bid":
		var data kurir_pengiriman_services.PayloadAktifkanBidKurir
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_pengiriman_services.AktifkanBidKurir(ctx, data, db)
	case "/kurir/pengiriman/update-posisi-bid":
		var data kurir_pengiriman_services.PayloadUpdatePosisiBid
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = kurir_pengiriman_services.UpdatePosisiBidKurir(ctx, data, db)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(hasil)
}
