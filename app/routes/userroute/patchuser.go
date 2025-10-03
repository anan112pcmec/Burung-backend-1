package userroute

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	pengguna_service "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services"
	pengguna_credential_services "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/credential_services"
	pengguna_profiling_services "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/profiling_services"
	pengguna_social_media_service "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/social_media_services"
	pengguna_transaction_services "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/transaction_services"

)

func PatchUserHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, rds_barang *redis.Client, rds_engagement *redis.Client) {
	var hasil *response.ResponseForm
	ctx := r.Context()

	switch r.URL.Path {
	case "/user/likes-barang":
		var data pengguna_service.PayloadLikesBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.LikesBarang(data, db, rds_barang)
	case "/user/komentar-barang/edit":
		var data pengguna_service.PayloadEditKomentarBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.EditKomentarBarang(ctx, data, db)
	case "/user/keranjang-barang/edit":
		var data pengguna_service.PayloadEditDataKeranjangBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.EditKeranjangBarang(ctx, data, db)
	case "/user/profiling/personal-update":
		var data pengguna_profiling_services.PayloadPersonalProfilingPengguna
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_profiling_services.UbahPersonalProfilingPengguna(ctx, data, db)
	case "/user/credential/update-password":
		var data pengguna_credential_services.PayloadPreUbahPasswordPengguna
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_credential_services.PreUbahPasswordPengguna(data, db, rds_engagement)
	case "/user/credential/validate-password-otp":
		var data pengguna_credential_services.PayloadValidateOTPPasswordPengguna
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_credential_services.ValidateUbahPasswordPenggunaViaOtp(data, db, rds_engagement)
	case "/user/credential/validate-password-pin":
		var data pengguna_credential_services.PayloadValidatePinPasswordPengguna
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_credential_services.ValidateUbahPasswordPenggunaViaPin(data, db, rds_engagement)
	case "/user/credential/update-pin":
		var data pengguna_credential_services.PayloadUpdatePinPengguna
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_credential_services.UpdateSecretPinPengguna(data, db)
	case "/user/transaksi/payment-gateaway-snap/va":
		var data pengguna_transaction_services.PayloadSnapTransaksiRequest
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.SnapTransaksi(data, db)
	case "/user/transaksi/payment-gateawat-snap-berhasil":
		var data pengguna_transaction_services.PayloadLockTransaksi
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.LockTransaksi(data, db)
	case "/user/social-media/engage-social-media":
		var data pengguna_social_media_service.PayloadEngageSocialMedia
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_social_media_service.EngageSocialMediaPengguna(data, db)
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
