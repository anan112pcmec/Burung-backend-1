package seller_serve

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/database/models"
	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func AmbilRandomSeller(ctx context.Context, db *gorm.DB, rds *redis.Client) *response.ResponseForm {
	var wg sync.WaitGroup
	var mu sync.Mutex
	services := "AmbilRandomSeller"
	var sellers []models.Seller
	const maxSeller = 20

	keys, _ := rds.SRandMemberN(ctx, "barang_keys", int64(maxSeller)).Result()
	if len(keys) == 0 {
		return &response.ResponseForm{
			Status:   http.StatusNotFound,
			Services: services,
			Payload:  "Gagal, mungkin server sedang sibuk atau maintenance. Coba lagi nanti.",
		}
	}

	if len(keys) > 0 {
		for _, key := range keys {
			if key == "_init_" {
				continue
			}

			wg.Add(1)
			go func(key string) {
				defer wg.Done()
				var seller models.Seller

				data, _ := rds.HGetAll(ctx, key).Result()

				id_seller, err_id := strconv.Atoi(data["id"])
				if err_id != nil {
					return
				}

				follower, err_fol := strconv.Atoi(data["follower_total_seller"])
				if err_fol != nil {
					return
				}

				if id_seller != 0 {
					seller = models.Seller{
						ID:               int32(id_seller),
						Username:         data["username_seller"],
						Nama:             data["nama_seller"],
						JamOperasional:   data["jam_operasional_seller"],
						SellerDedication: data["seller_dedication_seller"],
						Deskripsi:        data["deskripsi_seller"],
						FollowerTotal:    int32(follower),
						Email:            data["email_seller"],
						Jenis:            data["jenis_seller"],
						Punchline:        data["punchline_seller"],
					}

					mu.Lock()
					sellers = append(sellers, seller)
					mu.Unlock()
				}

			}(key)
		}

		wg.Wait()
	} else {
		if err_db := db.Model(&models.Seller{}).Limit(maxSeller).Find(&sellers).Error; err_db != nil {
			return &response.ResponseForm{
				Status:   http.StatusNotFound,
				Services: services,
				Payload:  "Gagal, mungkin server sedang sibuk atau maintenance. Coba lagi nanti.",
			}
		}
	}

	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload:  sellers,
	}
}

// func AmbilSellerByNama(ctx context.Context, nama_seller string, db *gorm.DB, rds *redis.Client, SE meilisearch.ServiceManager) *response.ResponseForm {
// 	services := "AmbilSellerByNama"
// 	var sellers []models.Seller

// 	searchRes, _ := SE.Index("seller_all").Search(nama_seller, &meilisearch.SearchRequest{})

// 	if len(searchRes.Hits) > 0 {

// 	}
// 	return
// }
