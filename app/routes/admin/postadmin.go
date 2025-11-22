package admin_routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
)

func AdminPostHandler(w http.ResponseWriter, r *http.Request) {
	var hasil *response.ResponseForm

	switch r.URL.Path {
	case "/admin/":
		fmt.Println("Halo dari post")

		hasil = &response.ResponseForm{
			Status:   http.StatusOK,
			Services: "Test",
			Message:  "Halo dari admin endpoint",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(hasil.Status))
	json.NewEncoder(w).Encode(hasil)
}
