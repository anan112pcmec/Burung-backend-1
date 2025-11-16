package userroute

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	pengguna_alamat_services "github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/alamat_services"
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
	case "/user/barang/likes-barang":
		var data pengguna_service.PayloadLikesBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.LikesBarang(ctx, data, db)
	case "/user/barang/unlikes-barang":
		var data pengguna_service.PayloadUnlikeBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.UnlikeBarang(ctx, data, db)
	case "/user/barang/komentar-barang/edit":
		var data pengguna_service.PayloadEditKomentarBarangInduk
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.EditKomentarBarang(ctx, data, db)
	case "/user/barang/komentar-child/edit":
		var data pengguna_service.PayloadEditChildKomentar
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_service.EditChildKomentar(ctx, data, db)
	case "/user/barang/keranjang-barang/edit":
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
		hasil = pengguna_credential_services.PreUbahPasswordPengguna(ctx, data, db, rds_engagement)
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
		hasil = pengguna_credential_services.UpdateSecretPinPengguna(ctx, data, db)
	case "/user/transaksi/payment-gateaway-snap":
		var data pengguna_transaction_services.PayloadSnapTransaksiRequest
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.SnapTransaksi(ctx, data, db)
	case "/user/transaksi/payment-gateaway-snap-berhasil/va":
		var data pengguna_transaction_services.PayloadLockTransaksiVa
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.LockTransaksiVa(data, db)
	case "/user/transaksi/payment-gateaway-snap-paided-failed/va":
		var data pengguna_transaction_services.PayloadPaidFailedTransaksiVa
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.PaidFailedTransaksiVa(data, db)
	case "/user/transaksi/payment-gateaway-snap-berhasil/wallet":
		var data pengguna_transaction_services.PayloadLockTransaksiWallet
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.LockTransaksiWallet(data, db)
	case "/user/transaksi/payment-gateaway-snap-paided-failed/wallet":
		var data pengguna_transaction_services.PayloadPaidFailedTransaksiWallet
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.PaidFailedTransaksiWallet(data, db)
	case "/user/transaksi/payment-gateaway-snap-berhasil/gerai":
		var data pengguna_transaction_services.PayloadLockTransaksiGerai
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.LockTransaksiGerai(data, db)
	case "/user/transaksi/payment-gateaway-snap-paided-failed/gerai":
		var data pengguna_transaction_services.PayloadPaidFailedTransaksiGerai
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_transaction_services.PaidFailedTransaksiGerai(data, db)

	// case "/user/transaksi/payment-gateaway-snap-pending":
	// 	var data pengguna_transaction_services.PayloadPendingTransaksi
	// 	if err := helper.DecodeJSONBody(r, &data); err != nil {
	// 		http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// 	hasil = pengguna_transaction_services.PendingTransaksi(ctx, data, db, rds_engagement)
	case "/user/social-media/engage-social-media":
		var data pengguna_social_media_service.PayloadEngageTautkanSocialMedia
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_social_media_service.EngageTautkanSocialMediaPengguna(ctx, data, db)
	case "/user/alamat/edit-alamat":
		var data pengguna_alamat_services.PayloadEditAlamatPengguna
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = pengguna_alamat_services.EditAlamatPengguna(ctx, data, db)
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
