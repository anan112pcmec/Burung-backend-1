package userroute

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/data-serve/barang_serve"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func GetUserHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, rds *redis.Client, SE meilisearch.ServiceManager) {
	var hasil *response.ResponseForm
	ctx := r.Context()

	switch r.URL.Path {
	case "/user/barang-all":
		hasil = barang_serve.AmbilRandomBarang(ctx, rds)
	case "/user/barang-spesified":
		jenis := r.URL.Query().Get("jenis")
		if jenis != "" {
			hasil = barang_serve.AmbilBarangJenis(ctx, db, rds, jenis)
		}

		seller := r.URL.Query().Get("seller")
		if seller != "" {
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangSeller(ctx, rds, int32(seller_id))
		}

		nama_barang := r.URL.Query().Get("nama_barang")
		if nama_barang != "" {
			hasil = barang_serve.AmbilBarangNama(ctx, rds, db, nama_barang, SE)
		}

		if nama_barang != "" && jenis != "" {
			hasil = barang_serve.AmbilBarangNamaDanJenis(ctx, rds, db, nama_barang, jenis, SE)
		}

		if nama_barang != "" && seller != "" {
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangNamaDanSeller(ctx, rds, db, int32(seller_id), nama_barang, SE)
		}

		if jenis != "" && seller != "" {

		}
	case "/user/data-barang-induk":
		id_barang_induk, err := strconv.Atoi(r.URL.Query().Get("barang_induk"))
		hasil = barang_serve.AmbilDataBarangInduk(ctx, int32(id_barang_induk), db, rds)
		if err != nil {
			hasil = &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: "User Services",
				Payload:  "Barang Itu Tidak Ada",
			}
		}
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
