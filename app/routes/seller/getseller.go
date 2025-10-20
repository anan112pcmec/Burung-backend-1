package seller

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

func GetSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request, rds *redis.Client, SE meilisearch.ServiceManager) {
	var hasil *response.ResponseForm
	ctx := r.Context()

	switch r.URL.Path {
	case "/seller/barang-all":
		FinalTake, err := strconv.Atoi(r.URL.Query().Get("finalTake"))
		if err != nil {
			hasil = &response.ResponseForm{
				Status: http.StatusUnauthorized,
			}
			return
		}

		hasil = barang_serve.AmbilRandomBarang(FinalTake, ctx, rds, db)
	case "/seller/barang-spesified":
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
		case nama_barang != "" && jenis != "":
			hasil = barang_serve.AmbilBarangNamaDanJenis(FinalTake, ctx, rds, db, nama_barang, jenis, SE)

		case nama_barang != "" && seller != "":
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangNamaDanSeller(FinalTake, ctx, rds, db, int32(seller_id), nama_barang, SE)

		case jenis != "":
			hasil = barang_serve.AmbilBarangJenis(FinalTake, ctx, db, rds, jenis)

		case seller != "":
			seller_id, _ := strconv.Atoi(seller)
			hasil = barang_serve.AmbilBarangSeller(FinalTake, ctx, db, rds, int32(seller_id))

		case nama_barang != "":
			hasil = barang_serve.AmbilBarangNama(FinalTake, ctx, rds, db, nama_barang, SE)

		default:
			hasil = &response.ResponseForm{
				Status:  http.StatusBadRequest,
				Payload: "Parameter tidak valid",
			}
		}

	case "/seller/data-barang-induk":
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
