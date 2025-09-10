package pengguna_service

import (
	"net/http"

	"github.com/redis/go-redis/v9"

	"github.com/anan112pcmec/Burung-backend-1/app/response"
	"github.com/anan112pcmec/Burung-backend-1/app/service/pengguna_service/barang_services/response_barang_user"



)

func AmbilRandomBarang(rds redis.Client) *response.ResponseForm {
	services := "AmbilRandomBarang"

	barang := make(chan [30]map[string]response_barang_user.UserBarang{

	})


	return &response.ResponseForm{
		Status:   http.StatusOK,
		Services: services,
		Payload: response_barang_user.ResponseUserBarang{
			Barang: ,
		},
	}
}
