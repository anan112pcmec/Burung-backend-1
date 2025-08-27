package userroute

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

func GetUserHandler(db *gorm.DB, w http.ResponseWriter, r *http.Request) {

	var hasil any

	fmt.Println("dari /user ini jalan")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  200,
		"message": "Halo dari GET middleware handler",
		"payload": hasil,
	})
}
