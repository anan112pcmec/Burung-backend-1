package seller

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_alamat_services "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/alamat_services"
	seller_service "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services"
	seller_credential_services "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/credential_services"
	"github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/jenis_seller_services"
	seller_order_processing_services "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/order_processing_services"
	seller_profiling_services "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/profiling_services"
)

func PatchSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, rds_engagement *redis.Client) {
	var hasil *response.ResponseForm

	ctx := r.Context()

	switch r.URL.Path {
	case "/seller/edit_barang":
		var data seller_service.PayloadEditBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.EditBarang(db, data)
	case "/seller/edit_kategori_barang":
		var data seller_service.PayloadEditKategori
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.EditKategoriBarang(db, data)
	case "/seller/edit/stok-barang":
		var data seller_service.PayloadEditStokBarang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.EditStokBarang(db, data)
	case "/seller/profiling/personal-update":
		var data seller_profiling_services.PayloadUpdateProfilePersonalSeller
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_profiling_services.UpdatePersonalSeller(ctx, db, data)
	case "/seller/profiling/info-general-update":
		var data seller_profiling_services.PayloadUpdateInfoGeneralSeller
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_profiling_services.UpdateInfoGeneralPublic(db, data)
	case "/seller/credential/update-password":
		var data seller_credential_services.PayloadPreUbahPasswordSeller
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_credential_services.PreUbahPasswordSeller(data, db, rds_engagement)
	case "/seller/credential/validate-password-otp":
		var data seller_credential_services.PayloadValidateUbahPasswordSellerOTP
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_credential_services.ValidateUbahPasswordSeller(data, db, rds_engagement)
	case "/seller/order-processing/approve":
		var data seller_order_processing_services.PayloadApproveOrder
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_order_processing_services.ApproveOrderBarang(data, db)
	case "/seller/order-processing/unapprove":
		var data seller_order_processing_services.PayloadUnApproveOrder
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_order_processing_services.UnApproveOrderBarang(data, db)
	case "/seller/jenis/ajukan-perubahan":
		var data jenis_seller_services.PayloadAjukanUbahJenisSeller
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = jenis_seller_services.AjukanUbahJenisSeller(data, db)
	case "/seller/alamat/edit-alamat-gudang":
		var data seller_alamat_services.PayloadEditAlamatGudang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_alamat_services.EditAlamatGudang(data, db)
	case "/seller/barang/edit-alamat-barang-induk":
		var data seller_service.PayloadEditAlamatBarangInduk
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.EditAlamatGudangBarangInduk(data, db)
	case "/seller/barang/edit-alamat-barang-kategori":
		var data seller_service.PayloadEditAlamatBarangKategori
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.EditAlamatGudangBarangKategori(data, db)
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
