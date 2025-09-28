package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/routes/auth"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/kurir"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/seller"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/userroute"
)

func PostHandler(db *gorm.DB, rds *redis.Client, rds_engagement *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("PostHandler dijalankan...")

		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/auth/" {
			fmt.Println("Auth HANDLER JALAN")
			auth.HandleAuth(db, w, r, rds)
			return
		}

		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/user/" {
			userroute.PostUserHandler(db, w, r, rds_engagement)
			return
		}

		if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/seller/" {
			seller.PostSellerHandler(db, w, r)
			return
		}

		// Jika path diawali "/kurir/"
		if len(r.URL.Path) >= 7 && r.URL.Path[:7] == "/kurir/" {
			kurir.PostKurirHandler(db, w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "url mu tidak jelas",
		})
	}
}
