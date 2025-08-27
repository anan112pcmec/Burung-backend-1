package seller

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

func PostSellerHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	var hasil any

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "dari GET middleware handler",
		"payload": hasil,
	})
}
