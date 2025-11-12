package seller

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/helper"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
	seller_alamat_services "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/alamat_services"
	seller_service "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/barang_services"
	seller_credential_services "github.com/anan112pcmec/Burung-backend-1/app/service/seller_services/credential_services"
)

func DeleteSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm

	ctx := r.Context()

	switch r.URL.Path {
	case "/seller/hapus_barang":
		var data seller_service.PayloadHapusBarangInduk
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
		}
		hasil = seller_service.HapusBarangInduk(ctx, db, data)
	case "/seller/hapus_kategori_barang":
		var data seller_service.PayloadHapusKategori
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
		}
		hasil = seller_service.HapusKategoriBarang(ctx, db, data)
	case "/seller/credential/hapus-rekening":
		var data seller_credential_services.PayloadHapusNorekSeller
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
		}
		hasil = seller_credential_services.HapusRekeningSeller(ctx, data, db)
	case "/seller/alamat/hapus-alamat-gudang":
		var data seller_alamat_services.PayloadHapusAlamatGudang
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_alamat_services.HapusAlamatGudang(ctx, data, db)
	case "/seller/komentar-barang/hapus":
		var data seller_service.PayloadHapusKomentarBarangInduk
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.HapusKomentarBarang(ctx, data, db)
	case "/seller/komentar-child/hapus":
		var data seller_service.PayloadHapusChildKomentar
		if err := helper.DecodeJSONBody(r, &data); err != nil {
			http.Error(w, "Gagal parsing JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		hasil = seller_service.HapusChildKomentar(ctx, data, db)
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
