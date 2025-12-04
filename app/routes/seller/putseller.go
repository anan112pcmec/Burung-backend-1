package seller

import (
	"encoding/json"
	"net/http"

	"github.com/anan112pcmec/Burung-backend-1/app/config"
)

func PutSellerHandler(db *config.InternalDBReadWriteSystem, w http.ResponseWriter, r *http.Request) {
	var hasil any

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "dari GET middleware handler",
		"payload": hasil,
	})
}
