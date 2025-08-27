package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func DeleteHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("PostHandler dijalankan...")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Data POST diproses di middleware",
		})
	}
}
