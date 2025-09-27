package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/routes/kurir"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/seller"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/userroute"
)

func PatchHandler(db *gorm.DB, rds_barang *redis.Client, rds_engagement *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("PatchHandler dijalankan...")

		// Jika path diawali "/user/"
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/user/" {
			userroute.PatchUserHandler(db, w, r, rds_barang, rds_engagement)
			return
		}

		// Jika path diawali "/seller/"
		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/seller/" {
			seller.PatchSellerHandler(db, w, r, rds_engagement)
			return
		}

		if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/kurir/" {
			kurir.PatchKurirHandler(db, w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "url mu tidak jelas",
		})
	}
}
