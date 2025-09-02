package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/routes/auth"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/kurir"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/seller"
	"github.com/anan112pcmec/Burung-backend-1/app/routes/userroute"
)

func PostHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("PostHandler dijalankan...")

		//registration handler
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/auth/" {
			fmt.Println("Auth HANDLER JALAN")
			auth.HandleAuth(db, w, r)
			return
		}

		// Jika path diawali "/user/"
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/user/" {
			userroute.PostUserHandler(db, w, r)
			return
		}

		// Jika path diawali "/seller/"
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
