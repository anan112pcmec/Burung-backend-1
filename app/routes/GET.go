package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/routes/kurir"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/seller"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/userroute"
)

func GetHandler(db *gorm.DB, redis_barang *redis.Client, redis_entity *redis.Client, SE meilisearch.ServiceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GetHandler dijalankan...")

		// Jika path diawali "/user/"
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/user/" {
			userroute.GetUserHandler(db, w, r, redis_barang, redis_entity, SE)
			return
		}

		// Jika path diawali "/seller/"
		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/seller/" {
			seller.GetSellerHandler(db, w, r, redis_barang, SE)
			return
		}

		// Jika path diawali "/kurir/"
		if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/kurir/" {
			kurir.GetKurirHandler(db, w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "url mu tidak jelas",
		})
	}
}
