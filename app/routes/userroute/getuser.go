package userroute

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
	case "/user/barang-spesified":
		FinalTake, err := strconv.Atoi(r.URL.Query().Get("finalTake"))
		if err != nil {
			hasil = &response.ResponseForm{
				Status: http.StatusUnauthorized,
			}
			return
		}
		jenis := r.URL.Query().Get("jenis")
		seller := r.URL.Query().Get("seller")
		nama_barang := r.URL.Query().Get("nama_barang")

		switch {
		// --- Kombinasi 3 parameter ---
		// case nama_barang != "" && jenis != "" && seller != "":
		// 	seller_id, _ := strconv.Atoi(seller)
		// 	hasil = barang_serve.AmbilBarangnama(ctx, redis_barang, db, int32(seller_id), nama_barang, jenis, SE)

		// --- Kombinasi 2 parameter ---
		case nama_barang != "" && jenis != "":
			hasil = barang_serve.AmbilBarangNamaDanJenis(FinalTake, ctx, redis_barang, db, nama_barang, jenis, SE)

		case nama_barang != "" && seller != "":
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangNamaDanSeller(FinalTake, ctx, redis_barang, db, int32(seller_id), nama_barang, SE)

		case jenis != "" && seller != "":
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangJenisDanSeller(FinalTake, ctx, redis_barang, db, int32(seller_id), jenis)

		// --- Hanya 1 parameter ---
		case nama_barang != "":
			hasil = barang_serve.AmbilBarangNama(FinalTake, ctx, redis_barang, db, nama_barang, SE)

		case jenis != "":
			hasil = barang_serve.AmbilBarangJenis(FinalTake, ctx, db, redis_barang, jenis)

		case seller != "":
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangSeller(FinalTake, ctx, db, redis_barang, int32(seller_id))

		// --- Tidak ada parameter valid ---
		default:
			hasil = &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: "Barang Services",
				Payload:  "Parameter pencarian tidak lengkap. Gunakan salah satu dari: nama_barang, jenis, seller (atau kombinasi yang valid).",
			}
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
		finalTake := r.URL.Query().Get("finalTake")
		finalNumber, err := strconv.Atoi(finalTake)
		if err != nil {
			hasil = &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: "Seller Services",
				Payload:  "Gagal Coba Lagi Nanti",
			}
		}
		hasil = seller_serve.AmbilRandomSeller(int64(finalNumber), ctx, db, redis_entity)
	case "/user/data-seller-spesified":
		finalTake := r.URL.Query().Get("finalTake")
		finalNumber, err := strconv.Atoi(finalTake)
		if err != nil {
			hasil = &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: "Seller Services",
				Payload:  "Gagal Coba Lagi Nanti",
			}
		}

		nama := strings.TrimSpace(r.URL.Query().Get("nama"))
		jenis := strings.TrimSpace(r.URL.Query().Get("jenis"))
		dedication := strings.TrimSpace(r.URL.Query().Get("dedication"))

		switch {
		case nama != "" && jenis != "" && dedication != "":
			hasil = seller_serve.AmbilSellerByNamaJenisDedication(int64(finalNumber), ctx, nama, jenis, dedication, db, redis_entity, SE)
		case nama != "" && jenis != "":
			hasil = seller_serve.AmbilSellerByNamaDanJenis(int64(finalNumber), ctx, nama, jenis, db, redis_entity, SE)
		case nama != "" && dedication != "":
			hasil = seller_serve.AmbilSellerByNamaDanDedication(int64(finalNumber), ctx, nama, dedication, db, redis_entity, SE)
		case jenis != "" && dedication != "":
			hasil = seller_serve.AmbilSellerByJenisDanDedication(int64(finalNumber), ctx, jenis, dedication, db, redis_entity)
		case nama != "":
			hasil = seller_serve.AmbilSellerByNama(int64(finalNumber), ctx, nama, db, redis_entity, SE)
		case jenis != "":
			hasil = seller_serve.AmbilSellerByJenis(int64(finalNumber), ctx, jenis, db, redis_entity)
		case dedication != "":
			hasil = seller_serve.AmbilSellerByDedication(int64(finalNumber), ctx, dedication, db, redis_entity)
		default:
			hasil = &response.ResponseForm{
				Status:   http.StatusBadRequest,
				Services: "Seller Services",
				Payload:  "Parameter pencarian tidak lengkap. Gunakan salah satu dari: nama, jenis, dedication (atau kombinasi yang valid).",
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
