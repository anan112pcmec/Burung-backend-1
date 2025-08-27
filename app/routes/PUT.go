package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"

	"github.com/anan112pcmec/Burung-backend-1/app/routes/userroute"
)

func PutHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("PutHandler dijalankan...")

		// Jika path diawali "/user/"
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/user/" {
			userroute.PutUserHandler(db, w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "url mu tidak jelas",
		})
	}
}
