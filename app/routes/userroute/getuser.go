package userroute

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/data-serve/barang_serve"
	"github.com/anan112pcmec/Burung-backend-1/app/data-serve/seller_serve"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func GetUserHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, redis_barang *redis.Client, redis_entity *redis.Client, SE meilisearch.ServiceManager) {
	var hasil *response.ResponseForm
	ctx := r.Context()

	switch r.URL.Path {
	case "/user/barang-all":
		hasil = barang_serve.AmbilRandomBarang(ctx, redis_barang)
	case "/user/barang-spesified":
		jenis := r.URL.Query().Get("jenis")
		if jenis != "" {
			hasil = barang_serve.AmbilBarangJenis(ctx, db, redis_barang, jenis)
		}

		seller := r.URL.Query().Get("seller")
		if seller != "" {
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangSeller(ctx, redis_barang, int32(seller_id))
		}

		nama_barang := r.URL.Query().Get("nama_barang")
		if nama_barang != "" {
			hasil = barang_serve.AmbilBarangNama(ctx, redis_barang, db, nama_barang, SE)
		}

		if nama_barang != "" && jenis != "" {
			hasil = barang_serve.AmbilBarangNamaDanJenis(ctx, redis_barang, db, nama_barang, jenis, SE)
		}

		if nama_barang != "" && seller != "" {
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangNamaDanSeller(ctx, redis_barang, db, int32(seller_id), nama_barang, SE)
		}

		if jenis != "" && seller != "" {

		}
	case "/user/data-barang-induk":
		id_barang_induk, err := strconv.Atoi(r.URL.Query().Get("barang_induk"))
		hasil = barang_serve.AmbilDataBarangInduk(ctx, int32(id_barang_induk), db, redis_barang)
		if err != nil {
			hasil = &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: "User Services",
				Payload:  "Barang Itu Tidak Ada",
			}
		}
	case "/user/data-seller-all":
		hasil = seller_serve.AmbilRandomSeller(ctx, db, redis_entity)
	case "/user/data-seller-spesified":
		nama := r.URL.Query().Get("nama")
		if nama != "" {
			hasil = seller_serve.AmbilSellerByNama(ctx, nama, db, redis_entity, SE)
		}

		jenis := r.URL.Query().Get("jenis")
		if jenis != "" {
			hasil = seller_serve.AmbilSellerByJenis(ctx, jenis, db, redis_entity)
		}

		if nama != "" && jenis != "" {
			hasil = seller_serve.AmbilSellerByNamaDanJenis(ctx, nama, jenis, db, redis_entity, SE)
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
