package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"

	"github.com/anan112pcmec/Burung-backend-1/app/config"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/kurir"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/seller"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/userroute"
)

func PatchHandler(db *config.InternalDBReadWriteSystem, rds_barang *redis.Client, rds_engagement *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("PatchHandler dijalankan...")

		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/user/" {
			userroute.PatchUserHandler(db, w, r, rds_barang, rds_engagement)
			return
		}

		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/seller/" {
			seller.PatchSellerHandler(db, w, r, rds_engagement)
			return
		}

		if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/kurir/" {
			kurir.PatchKurirHandler(db, w, r, rds_engagement)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "url mu tidak jelas",
		})
	}
}
